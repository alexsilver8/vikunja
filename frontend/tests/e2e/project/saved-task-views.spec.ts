import {type APIRequestContext, type Page} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {createProjects} from './prepareProjects'

// The trigger is labelled with the active saved view's title, or "Saved views" when none matches.
function trigger(page: Page, name: string) {
	return page.locator('.filter-container').getByRole('button', {name, exact: true})
}

async function storedViews(apiContext: APIRequestContext, token: string, projectId: number, viewId: number) {
	const response = await apiContext.get(`/api/v2/projects/${projectId}/views/${viewId}/saved-views`, {
		headers: {Authorization: `Bearer ${token}`},
	})
	expect(response.ok()).toBe(true)
	return (await response.json()).items
}

test.describe('Saved task views', () => {
	test('List view: save, apply, rename and delete a view', async ({authenticatedPage: page, apiContext, userToken}) => {
		const [project] = await createProjects(1)
		const listView = project.views[0].id
		await TaskFactory.create(3, {
			id: '{increment}',
			project_id: project.id,
			title: (i: number) => `Task ${i}`,
			priority: (i: number) => i,
			due_date: (i: number) => new Date(2030, 0, 4 - i).toISOString(),
		})

		await page.goto(`/projects/${project.id}/${listView}?filter=${encodeURIComponent('priority >= 2')}&sort=due_date:asc`)
		await expect(page.locator('.tasks .task .tasktext')).toHaveCount(2)
		await expect(page.locator('.tasks .task .tasktext')).toContainText(['Task 3', 'Task 2'])

		await trigger(page, 'Saved views').click()
		await expect(page.getByText('No saved views yet')).toBeVisible()
		await page.getByLabel('Name of the current view').fill('Important')
		await page.getByRole('button', {name: 'Save current view'}).click()
		await expect(page.getByRole('button', {name: 'Apply saved view Important'})).toBeVisible()
		// The saved view matches what is on screen, so the trigger shows its name.
		await expect(trigger(page, 'Important')).toBeVisible()

		const [saved] = await storedViews(apiContext, userToken, project.id, listView)
		expect(saved).toMatchObject({
			title: 'Important',
			filter: 'priority >= 2',
			search: '',
			sort_by: ['due_date'],
			order_by: ['asc'],
		})

		await page.goto(`/projects/${project.id}/${listView}?sort=title:desc`)
		await expect(page.locator('.tasks .task .tasktext')).toHaveCount(3)
		await expect(page.locator('.tasks .task .tasktext')).toContainText(['Task 3', 'Task 2', 'Task 1'])
		await trigger(page, 'Saved views').click()
		await page.getByRole('button', {name: 'Apply saved view Important'}).click()
		await expect(page).toHaveURL(/sort=due_date:asc/)
		await expect(page).toHaveURL(/filter=priority/)
		await expect(page.locator('.tasks .task .tasktext')).toHaveCount(2)
		await expect(page.locator('.tasks .task .tasktext')).toContainText(['Task 3', 'Task 2'])

		await page.reload()
		await trigger(page, 'Important').click()
		await page.getByRole('button', {name: 'Rename saved view Important'}).click()
		const renameInput = page.getByLabel('New name for Important')
		await expect(renameInput).toBeFocused()
		await renameInput.fill('Urgent')
		await renameInput.press('Enter')
		await expect(page.getByRole('button', {name: 'Apply saved view Urgent'})).toBeVisible()
		expect((await storedViews(apiContext, userToken, project.id, listView))[0].title).toBe('Urgent')

		await page.getByRole('button', {name: 'Delete saved view Urgent'}).click()
		await expect(page.getByText('No saved views yet')).toBeVisible()
		expect(await storedViews(apiContext, userToken, project.id, listView)).toEqual([])
	})

	test('Table view: applying a view updates the column sort', async ({authenticatedPage: page, apiContext, userToken}) => {
		const [project] = await createProjects(1)
		const tableView = project.views[2].id
		await TaskFactory.create(3, {
			id: '{increment}',
			project_id: project.id,
			title: (i: number) => `Task ${i}`,
			due_date: (i: number) => new Date(2030, 0, i).toISOString(),
		})
		const created = await apiContext.post(`/api/v2/projects/${project.id}/views/${tableView}/saved-views`, {
			headers: {Authorization: `Bearer ${userToken}`},
			data: {
				title: 'By due date',
				sort_by: ['due_date'],
				order_by: ['desc'],
			},
		})
		expect(created.ok()).toBe(true)

		await page.goto(`/projects/${project.id}/${tableView}`)
		await trigger(page, 'Saved views').click()
		await page.getByRole('button', {name: 'Apply saved view By due date'}).click()

		await expect(page).toHaveURL(/sort=due_date:desc/)
		await expect(page.locator('th', {hasText: 'Due date'})).toHaveAttribute('aria-sort', 'descending')
		await expect(page.locator('.project-table tbody tr td:nth-child(3)')).toHaveText(['Task 3', 'Task 2', 'Task 1'])
	})
})
