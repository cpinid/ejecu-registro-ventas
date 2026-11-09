import React from "react";
import ReactDOM from "react-dom/client";

import RegistroVenta from "./componentes/registro-venta/registro-venta";

import "./styles.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <RegistroVenta />
  </React.StrictMode>,
);