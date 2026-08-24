import { existsSync, rmSync } from "node:fs"
import { join } from "node:path"

const staleRoutes = [
  join(process.cwd(), "app", "(public)", "newUserRecharge"),
]

for (const routePath of staleRoutes) {
  if (!existsSync(routePath)) {
    continue
  }

  rmSync(routePath, { recursive: true, force: true })
  console.log(`[prebuild] Removed stale route: ${routePath}`)
}
