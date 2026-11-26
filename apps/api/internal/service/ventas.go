package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"time"
)

var (
	ErrNoExiste  = errors.New("venta o producto inexistente")
	ErrConflicto = errors.New("clave reutilizada con otros datos")
	ErrInvalido  = errors.New("datos de venta invalidos")
	ErrStock     = errors.New("stock insuficiente")
	ErrExterno   = errors.New("dependencia no disponible; reintente con la misma clave")
)

type Item struct {
	ProductoID string `json:"producto_id" format:"uuid"`
	Cantidad   int32  `json:"cantidad" minimum:"1" maximum:"1000000"`
}
type Solicitud struct {
	MedioPago string `json:"medio_pago" enum:"efectivo,debito,credito,transferencia"`
	Productos []Item `json:"productos" minItems:"1" maxItems:"100"`
}
type Detalle struct {
	ProductoID     string `json:"producto_id" format:"uuid"`
	Nombre         string `json:"nombre"`
	Cantidad       int32  `json:"cantidad"`
	PrecioUnitario int64  `json:"precio_unitario"`
	CostoUnitario  int64  `json:"costo_unitario"`
	Subtotal       int64  `json:"subtotal"`
}
type Venta struct {
	ID         string    `json:"id" format:"uuid"`
	VendedorID string    `json:"vendedor_id" format:"uuid"`
	Fecha      time.Time `json:"fecha"`
	MedioPago  string    `json:"medio_pago"`
	Estado     string    `json:"estado"`
	Total      int64     `json:"total"`
	Detalles   []Detalle `json:"detalles"`
	Huella     string    `json:"-"`
}
type Producto struct {
	Nombre   string `json:"nombre"`
	Precio   int64  `json:"precio_venta"`
	Costo    int64  `json:"costo"`
	Vendible bool   `json:"vendible"`
}
type Datos interface {
	PorClave(context.Context, string, string) (Venta, error)
	Crear(context.Context, Venta, string) (Venta, error)
	Estado(context.Context, Venta, string) (Venta, error)
	PorID(context.Context, string) (Venta, error)
	Listar(context.Context, int32, int32) ([]Venta, error)
}
type Productos interface {
	Obtener(context.Context, string) (Producto, error)
}
type Stock interface {
	Descontar(context.Context, string, []Item) error
}
type Servicio struct {
	Datos     Datos
	Productos Productos
	Stock     Stock
}

func (s Servicio) Crear(ctx context.Context, usuario, clave string, req Solicitud) (Venta, error) {
	if len(req.Productos) == 0 || len(req.Productos) > 100 {
		return Venta{}, ErrInvalido
	}
	switch req.MedioPago {
	case "efectivo", "debito", "credito", "transferencia":
	default:
		return Venta{}, ErrInvalido
	}
	agrupados := map[string]int64{}
	for _, i := range req.Productos {
		if i.Cantidad <= 0 || i.Cantidad > 1000000 {
			return Venta{}, ErrInvalido
		}
		agrupados[i.ProductoID] += int64(i.Cantidad)
		if agrupados[i.ProductoID] > 1000000 {
			return Venta{}, ErrInvalido
		}
	}
	req.Productos = make([]Item, 0, len(agrupados))
	for id, n := range agrupados {
		req.Productos = append(req.Productos, Item{id, int32(n)})
	}
	sort.Slice(req.Productos, func(i, j int) bool { return req.Productos[i].ProductoID < req.Productos[j].ProductoID })
	b, err := json.Marshal(req)
	if err != nil {
		return Venta{}, err
	}
	h := sha256.Sum256(b)
	huella := hex.EncodeToString(h[:])
	v, err := s.Datos.PorClave(ctx, usuario, clave)
	if errors.Is(err, ErrNoExiste) {
		v = Venta{VendedorID: usuario, MedioPago: req.MedioPago, Huella: huella, Detalles: make([]Detalle, 0, len(req.Productos))}
		for _, i := range req.Productos {
			p, e := s.Productos.Obtener(ctx, i.ProductoID)
			if e != nil {
				return Venta{}, e
			}
			if !p.Vendible || p.Precio < 0 || p.Costo < 0 || p.Precio > math.MaxInt64/int64(i.Cantidad) {
				return Venta{}, ErrInvalido
			}
			sub := p.Precio * int64(i.Cantidad)
			if v.Total > math.MaxInt64-sub {
				return Venta{}, ErrInvalido
			}
			v.Total += sub
			v.Detalles = append(v.Detalles, Detalle{i.ProductoID, p.Nombre, i.Cantidad, p.Precio, p.Costo, sub})
		}
		v, err = s.Datos.Crear(ctx, v, clave)
	}
	if err != nil {
		return Venta{}, err
	}
	if v.Huella != huella {
		return Venta{}, ErrConflicto
	}
	switch v.Estado {
	case "confirmada":
		return v, nil
	case "rechazada":
		return Venta{}, ErrStock
	}
	// La clave de Stock es estable incluso si se pierde la respuesta HTTP.
	err = s.Stock.Descontar(ctx, "venta-"+v.ID, req.Productos)
	if errors.Is(err, ErrStock) {
		_, e := s.Datos.Estado(ctx, v, "rechazada")
		if e != nil {
			return Venta{}, e
		}
		return Venta{}, ErrStock
	}
	if err != nil {
		return Venta{}, err
	}
	return s.Datos.Estado(ctx, v, "confirmada")
}
func (s Servicio) PorID(ctx context.Context, id string) (Venta, error) { return s.Datos.PorID(ctx, id) }
func (s Servicio) Listar(ctx context.Context, lim, off int32) ([]Venta, error) {
	return s.Datos.Listar(ctx, lim, off)
}
