<template>
	<Popup
		placement="bottom-end"
		:anchor="triggerEl"
	>
		<template #trigger="{toggle}">
			<XButton
				ref="trigger"
				variant="secondary"
				icon="layer-group"
				class="saved-views-trigger"
				:class="{'has-active-view': activeView}"
				:wrap="false"
				@click.prevent.stop="toggle()"
			>
				<span
					ref="triggerLabel"
					v-tooltip="labelTruncated ? activeView?.title : undefined"
					class="saved-views-trigger-label"
				>
					{{ activeView?.title ?? $t('project.savedViews.title') }}
				</span>
			</XButton>
		</template>
		<template #content="{close}">
			<Card class="saved-views-popup">
				<p
					v-if="savedViews.length === 0"
					class="has-text-grey is-size-7 mbe-2"
				>
					{{ $t('project.savedViews.empty') }}
				</p>
				<ul
					v-else
					class="saved-views mbe-3"
				>
					<li
						v-for="view in savedViews"
						:key="view.id"
						class="saved-view"
					>
						<form
							v-if="renaming && renaming.id === view.id"
							class="saved-view-rename"
							@submit.prevent="saveRename()"
							@keydown.esc.stop="renaming = null"
						>
							<input
								ref="renameInput"
								v-model="renaming.title"
								class="input is-small"
								:aria-label="$t('project.savedViews.renameLabel', {title: view.title})"
								maxlength="250"
							>
							<BaseButton
								type="submit"
								class="saved-view-action"
								:aria-label="$t('project.savedViews.renameSave')"
								:aria-disabled="updateMutation.isPending.value || undefined"
							>
								<Icon icon="check" />
							</BaseButton>
							<BaseButton
								class="saved-view-action"
								:aria-label="$t('project.savedViews.renameCancel')"
								@click="renaming = null"
							>
								<Icon icon="times" />
							</BaseButton>
						</form>
						<template v-else>
							<BaseButton
								class="saved-view-title"
								:class="{'is-active': view.id === activeView?.id}"
								:aria-current="view.id === activeView?.id || undefined"
								:aria-label="$t('project.savedViews.apply', {title: view.title})"
								@click="apply(view, close)"
							>
								{{ view.title }}
							</BaseButton>
							<BaseButton
								v-if="view.id !== activeView?.id"
								v-tooltip="$t('project.savedViews.updateTooltip')"
								class="saved-view-action"
								:aria-label="$t('project.savedViews.update', {title: view.title})"
								:aria-disabled="updateMutation.isPending.value || undefined"
								@click="overwrite(view)"
							>
								<Icon :icon="['far', 'save']" />
							</BaseButton>
							<BaseButton
								v-tooltip="$t('project.savedViews.renameTooltip')"
								class="saved-view-action"
								:aria-label="$t('project.savedViews.rename', {title: view.title})"
								@click="startRename(view)"
							>
								<Icon icon="pen" />
							</BaseButton>
							<BaseButton
								v-tooltip="$t('project.savedViews.deleteTooltip')"
								class="saved-view-action is-danger"
								:aria-label="$t('project.savedViews.delete', {title: view.title})"
								:aria-disabled="deleteMutation.isPending.value || undefined"
								@click="remove(view)"
							>
								<Icon icon="trash-alt" />
							</BaseButton>
						</template>
					</li>
				</ul>
				<form
					class="saved-view-create"
					@submit.prevent="save()"
				>
					<label
						class="label is-small"
						for="saved-view-title"
					>
						{{ $t('project.savedViews.nameLabel') }}
					</label>
					<input
						id="saved-view-title"
						v-model="newTitle"
						class="input"
						:class="{'is-danger': showTitleRequired}"
						:placeholder="$t('project.savedViews.namePlaceholder')"
						:aria-invalid="showTitleRequired || undefined"
						:aria-describedby="showTitleRequired ? 'saved-view-title-error' : undefined"
						maxlength="250"
						@input="showTitleRequired = false"
					>
					<p
						v-if="showTitleRequired"
						id="saved-view-title-error"
						class="help is-danger"
					>
						{{ $t('project.savedViews.titleRequired') }}
					</p>
					<XButton
						type="submit"
						class="mbs-2 is-fullwidth"
						:icon="['far', 'save']"
						:shadow="false"
						:aria-disabled="createMutation.isPending.value || undefined"
						:loading="createMutation.isPending.value"
					>
						{{ $t('project.savedViews.save') }}
					</XButton>
				</form>
			</Card>
		</template>
	</Popup>
</template>

<script setup lang="ts">
import {computed, nextTick, ref} from 'vue'
import {useResizeObserver} from '@vueuse/core'

import BaseButton from '@/components/base/BaseButton.vue'
import XButton from '@/components/input/Button.vue'
import Card from '@/components/misc/Card.vue'
import Popup from '@/components/misc/Popup.vue'
import {
	useCreateSavedTaskViewMutation,
	useDeleteSavedTaskViewMutation,
	useUpdateSavedTaskViewMutation,
	type SavedTaskViewResponse,
	type SavedTaskViewState,
} from '@/client/queries/savedTaskViews'
import {useSavedTaskViews} from '@/composables/useSavedTaskViews'

const props = defineProps<{
	projectId: number,
	viewId: number,
	current: SavedTaskViewState,
}>()

const emit = defineEmits<{
	'apply': [view: SavedTaskViewResponse],
}>()

const trigger = ref<InstanceType<typeof XButton> | null>(null)
const triggerEl = computed<HTMLElement | null>(() => trigger.value?.$el ?? null)

const triggerLabel = ref<HTMLElement | null>(null)
const labelTruncated = ref(false)
useResizeObserver(triggerLabel, () => {
	const label = triggerLabel.value
	labelTruncated.value = label !== null && label.scrollWidth > label.clientWidth
})

const scope = computed(() => ({
	projectId: props.projectId,
	viewId: props.viewId,
}))
const {savedViews} = useSavedTaskViews(scope)

const createMutation = useCreateSavedTaskViewMutation()
const updateMutation = useUpdateSavedTaskViewMutation()
const deleteMutation = useDeleteSavedTaskViewMutation()

function sameState(view: SavedTaskViewState, other: SavedTaskViewState) {
	return view.filter === other.filter &&
		view.search === other.search &&
		view.sort_by.join(',') === other.sort_by.join(',') &&
		view.order_by.join(',') === other.order_by.join(',')
}

const activeView = computed(() => savedViews.value.find(view => sameState(view, props.current)))

function apply(view: SavedTaskViewResponse, close: () => void) {
	emit('apply', view)
	close()
}

const newTitle = ref('')
const showTitleRequired = ref(false)

async function save() {
	if (createMutation.isPending.value) return
	const title = newTitle.value.trim()
	if (title === '') {
		showTitleRequired.value = true
		return
	}
	try {
		await createMutation.mutateAsync({
			scope: scope.value,
			body: {
				title,
				...props.current,
			},
		})
	} catch {
		return
	}
	if (newTitle.value.trim() === title) {
		newTitle.value = ''
	}
}

const renaming = ref<SavedTaskViewResponse | null>(null)
const renameInput = ref<HTMLInputElement[]>([])

async function startRename(view: SavedTaskViewResponse) {
	renaming.value = {...view}
	await nextTick()
	renameInput.value[0]?.focus()
}

async function saveRename() {
	if (!renaming.value || updateMutation.isPending.value) return
	const draft = renaming.value
	const title = draft.title.trim()
	if (title === '') return
	try {
		await updateMutation.mutateAsync({
			scope: scope.value,
			view: {
				...draft,
				title,
			},
		})
	} catch {
		return
	}
	if (renaming.value === draft) {
		renaming.value = null
	}
}

function overwrite(view: SavedTaskViewResponse) {
	if (updateMutation.isPending.value) return
	updateMutation.mutate({
		scope: scope.value,
		view: {
			...view,
			...props.current,
		},
	})
}

function remove(view: SavedTaskViewResponse) {
	if (deleteMutation.isPending.value) return
	deleteMutation.mutate({
		scope: scope.value,
		id: view.id,
	})
}
</script>

<style lang="scss" scoped>
.saved-views-popup {
	inline-size: 300px;
	max-inline-size: 100%;
	margin: 0;
}

.saved-views {
	list-style: none;
	margin: 0;
}

.saved-view {
	display: flex;
	align-items: center;
	gap: .25rem;
	border-radius: $radius;

	&:hover,
	&:focus-within {
		background: var(--grey-100);
	}
}

.saved-view-title {
	flex: 1;
	min-inline-size: 0;
	padding: .375rem .5rem;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
	text-align: start;

	&.is-active {
		font-weight: bold;
		color: var(--primary);
	}
}

.saved-view-action {
	padding: .375rem .5rem;
	color: var(--grey-500);

	&:hover,
	&:focus-visible {
		color: var(--grey-800);
	}

	&.is-danger:hover,
	&.is-danger:focus-visible {
		color: var(--danger);
	}
}

.saved-view-rename {
	display: flex;
	align-items: center;
	flex: 1;
	gap: .25rem;
	padding: .25rem;
}

.has-active-view {
	color: var(--primary);
}

// The button wraps its slot in an element that doesn't shrink, so the label caps itself.
.saved-views-trigger-label {
	display: inline-block;
	max-inline-size: 12rem;
	overflow: hidden;
	text-overflow: ellipsis;
	vertical-align: bottom;
}
</style>
