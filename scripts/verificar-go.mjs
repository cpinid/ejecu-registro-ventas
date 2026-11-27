import { spawnSync } from "node:child_process"
const fmt = spawnSync("gofmt", ["-l", "."], { cwd: "apps/api", encoding: "utf8" })
if (fmt.status !== 0 || fmt.stdout.trim()) {
  console.error(fmt.stdout, fmt.stderr)
  process.exit(1)
}
const vet = spawnSync("go", ["vet", "./..."], { cwd: "apps/api", stdio: "inherit" })
process.exit(vet.status ?? 1)
