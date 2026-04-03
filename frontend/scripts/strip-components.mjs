/**
 * Strip components/schemas from bundled OpenAPI YAML, keep securitySchemes.
 * The bundled spec duplicates schemas (inline in paths + in components),
 * which causes orval to generate duplicate TypeScript types.
 */
import { readFileSync, writeFileSync } from 'fs'

const file = '../spec/openapi-bundled.yaml'
const yaml = readFileSync(file, 'utf8')

// Find where components: starts
const componentsIdx = yaml.indexOf('\ncomponents:\n')
if (componentsIdx === -1) {
  console.log('No components section found, skipping')
  process.exit(0)
}

// Take everything before components
let result = yaml.substring(0, componentsIdx + 1)

// Re-add the security section (servers + security keys at root level come after components)
// Find servers: and security: sections after components
const afterComponents = yaml.substring(componentsIdx)
const serversMatch = afterComponents.match(/^servers:\n[\s\S]*?(?=^security:|\Z)/m)
const securityMatch = afterComponents.match(/^security:\n[\s\S]*/m)

// Add back components with only securitySchemes
result += `components:
  securitySchemes:
    ZitadelCookieAuth:
      type: apiKey
      in: cookie
      name: zitadel_access_token
    PlainCookieAuth:
      type: apiKey
      in: cookie
      name: sa_session
    BearerAuth:
      type: http
      scheme: bearer
`

if (serversMatch) result += '\n' + serversMatch[0]
if (securityMatch) result += '\n' + securityMatch[0]

writeFileSync(file, result)
console.log('Stripped components/schemas from bundled spec')
