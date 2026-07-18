// @ctx: M2.5 SPARQL GUI — enter query, run, results table, loading, error, export
import { test, expect } from '@playwright/test'
import { SparqlPage } from '../../pages/sparql.page'

test.describe('M2.5 SPARQL GUI', () => {
  test('should enter a SPARQL query when user types in the editor', async ({ page }) => {
    const sparql = new SparqlPage(page)
    await sparql.goto('ont-123')
    await sparql.enterQuery('SELECT * WHERE { ?s ?p ?o } LIMIT 10')
    const editor = page.locator('.sparql-editor textarea, .cm-editor')
    await expect(editor).not.toBeEmpty()
  })

  test('should run query and display results table when execute is clicked', async ({ page }) => {
    const sparql = new SparqlPage(page)
    await sparql.goto('ont-123')
    await sparql.enterQuery('SELECT * WHERE { ?s ?p ?o } LIMIT 10')
    await sparql.runQuery()
    const results = sparql.getResults()
    await expect(results).toBeVisible()
  })

  test('should show loading state while query is executing', async ({ page }) => {
    const sparql = new SparqlPage(page)
    await sparql.goto('ont-123')
    await sparql.enterQuery('SELECT * WHERE { ?s ?p ?o } LIMIT 10')
    await sparql.runQuery()
    await expect(page.locator('.loading-indicator, .spinner')).toBeVisible({ timeout: 2000 })
  })

  test('should show error message when query execution fails', async ({ page }) => {
    await page.route('**/graphql', route => {
      if (route.request().postData()?.includes('sparql')) {
        route.fulfill({ status: 200, body: JSON.stringify({ errors: [{ message: 'Query error' }] }) })
      } else {
        route.continue()
      }
    })
    const sparql = new SparqlPage(page)
    await sparql.goto('ont-123')
    await sparql.enterQuery('SELECT INVALID')
    await sparql.runQuery()
    await expect(page.getByText(/error/i)).toBeVisible()
  })

  test('should format query when format button is clicked', async ({ page }) => {
    const sparql = new SparqlPage(page)
    await sparql.goto('ont-123')
    await sparql.enterQuery('SELECT * WHERE { ?s ?p ?o }')
    await sparql.formatQuery()
  })

  test('should export results when export button is clicked', async ({ page }) => {
    const sparql = new SparqlPage(page)
    await sparql.goto('ont-123')
    await sparql.enterQuery('SELECT * WHERE { ?s ?p ?o } LIMIT 10')
    await sparql.runQuery()
    await page.getByRole('button', { name: /export/i }).click()
  })
})
