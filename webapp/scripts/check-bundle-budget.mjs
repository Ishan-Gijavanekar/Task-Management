import { gzipSync } from 'node:zlib'
import { readdirSync, readFileSync } from 'node:fs'
import { extname, join } from 'node:path'

const budgets = { '.js': 180 * 1024, '.css': 30 * 1024 }
const totals = { '.js': 0, '.css': 0 }
const assets = join(import.meta.dirname, '..', 'dist', 'assets')

for (const file of readdirSync(assets)) {
  const extension = extname(file)
  if (extension in totals)
    totals[extension] += gzipSync(readFileSync(join(assets, file))).length
}

for (const [extension, budget] of Object.entries(budgets)) {
  if (totals[extension] > budget) {
    throw new Error(
      `${extension} bundle is ${totals[extension]} bytes gzipped; budget is ${budget}`,
    )
  }
}

console.log(
  `Bundle budgets passed: JS ${totals['.js']} B, CSS ${totals['.css']} B (gzip)`,
)
