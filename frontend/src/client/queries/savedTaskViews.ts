import {queryOptions, useMutation} from '@tanstack/vue-query'

import {
	savedTaskViewsCreate,
	savedTaskViewsDelete,
	savedTaskViewsList,
	savedTaskViewsUpdate,
} from '@/client/generated'
import type {
	SavedTaskView,
	SavedTaskViewWritable,
} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {fetchAllPages} from './fetchAllPages'
import {i18n} from '@/i18n'

export type SavedTaskViewResponse = SavedTaskView & Required<Pick<SavedTaskView,
	'id' |
	'title' |
	'project_view_id' |
	'filter' |
	'search'
>> & {
	sort_by: string[]
	order_by: string[]
}

export type SavedTaskViewScope = {projectId: number, viewId: number}

export type SavedTaskViewState = Required<Pick<SavedTaskViewWritable, 'filter' | 'search'>> & {
	sort_by: string[]
	order_by: string[]
}

export function normalizeSavedTaskView(view: SavedTaskView): SavedTaskViewResponse {
	return {
		...view,
		id: view.id ?? 0,
		title: view.title ?? '',
		project_view_id: view.project_view_id ?? 0,
		filter: view.filter ?? '',
		search: view.search ?? '',
		sort_by: view.sort_by ?? [],
		order_by: view.order_by ?? [],
	}
}

export const savedTaskViewKeys = {
	list: ({projectId, viewId}: SavedTaskViewScope) => ['saved-task-views', projectId, viewId] as const,
}

export function savedTaskViewsQuery(scope: SavedTaskViewScope) {
	return queryOptions({
		queryKey: savedTaskViewKeys.list(scope),
		queryFn: async ({signal}) => (await fetchAllPages(async page => (await savedTaskViewsList({
			path: {
				project: scope.projectId,
				view: scope.viewId,
			},
			query: {page},
			signal,
		})).data)).map(normalizeSavedTaskView),
	})
}

function sortByTitle(views: SavedTaskViewResponse[]): SavedTaskViewResponse[] {
	return [...views].sort((a, b) => a.title.localeCompare(b.title, i18n.global.locale.value))
}

function path(scope: SavedTaskViewScope) {
	return {
		project: scope.projectId,
		view: scope.viewId,
	}
}

export function createSavedTaskViewMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({scope, body}: {scope: SavedTaskViewScope, body: SavedTaskViewWritable}) =>
			normalizeSavedTaskView((await savedTaskViewsCreate({path: path(scope), body})).data),
		onSuccess: (created, {scope}, client) => client.setQueryData<SavedTaskViewResponse[]>(
			savedTaskViewKeys.list(scope),
			current => current ? sortByTitle([...current, created]) : current,
		),
		onSettled: ({scope}, client) => client.invalidateQueries({queryKey: savedTaskViewKeys.list(scope)}),
		successMessage: () => i18n.global.t('project.savedViews.saveSuccess'),
	})
}

export function updateSavedTaskViewMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({scope, view}: {scope: SavedTaskViewScope, view: SavedTaskViewResponse}) =>
			normalizeSavedTaskView((await savedTaskViewsUpdate({
				path: {
					...path(scope),
					id: view.id,
				},
				body: {
					title: view.title,
					filter: view.filter,
					search: view.search,
					sort_by: view.sort_by,
					order_by: view.order_by,
				},
			})).data),
		onSuccess: (updated, {scope}, client) => client.setQueryData<SavedTaskViewResponse[]>(
			savedTaskViewKeys.list(scope),
			current => current ? sortByTitle(current.map(v => v.id === updated.id ? updated : v)) : current,
		),
		onSettled: ({scope}, client) => client.invalidateQueries({queryKey: savedTaskViewKeys.list(scope)}),
		successMessage: () => i18n.global.t('project.savedViews.renameSuccess'),
	})
}

export function deleteSavedTaskViewMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({scope, id}: {scope: SavedTaskViewScope, id: number}) =>
			(await savedTaskViewsDelete({
				path: {
					...path(scope),
					id,
				},
			})).data,
		onSuccess: (_data, {scope, id}, client) => client.setQueryData<SavedTaskViewResponse[]>(
			savedTaskViewKeys.list(scope),
			current => current?.filter(v => v.id !== id),
		),
		onSettled: ({scope}, client) => client.invalidateQueries({queryKey: savedTaskViewKeys.list(scope)}),
		successMessage: () => i18n.global.t('project.savedViews.deleteSuccess'),
	})
}

export function useCreateSavedTaskViewMutation() {
	return useMutation(createSavedTaskViewMutationOptions())
}

export function useUpdateSavedTaskViewMutation() {
	return useMutation(updateSavedTaskViewMutationOptions())
}

export function useDeleteSavedTaskViewMutation() {
	return useMutation(deleteSavedTaskViewMutationOptions())
}
