import { spawnSync } from "node:child_process"
import { writeFileSync, mkdirSync } from "node:fs"
const r = spawnSync("go", ["run", "./cmd/api", "openapi"], {
  cwd: "apps/api",
  encoding: "utf8",
})
if (r.status !== 0) throw Error(r.stderr || "Error al generar contrato")
mkdirSync("contracts", { recursive: true })
writeFileSync("contracts/openapi.yaml", r.stdout)
