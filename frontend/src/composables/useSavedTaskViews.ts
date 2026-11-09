import {computed, type MaybeRefOrGetter, toValue} from 'vue'
import {useQuery} from '@tanstack/vue-query'

import {savedTaskViewsQuery, type SavedTaskViewScope} from '@/client/queries/savedTaskViews'

export function useSavedTaskViews(scope: MaybeRefOrGetter<SavedTaskViewScope>) {
	const query = useQuery(computed(() => savedTaskViewsQuery(toValue(scope))))

	return {
		savedViews: computed(() => query.data.value ?? []),
		isPending: query.isPending,
	}
}
