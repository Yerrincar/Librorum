<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

export type FantasySelectOption = {
  label: string
  value: string
}

const props = defineProps<{
  id: string
  label: string
  modelValue: string
  name: string
  options: FantasySelectOption[]
  layout?: 'inline' | 'stacked'
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  change: []
}>()

const open = ref(false)
const root = ref<HTMLElement>()

const selectedLabel = computed(() => {
  return props.options.find((option) => option.value === props.modelValue)?.label ?? props.options[0]?.label ?? ''
})

onMounted(() => {
  document.addEventListener('click', handleDocumentClick)
  window.addEventListener('keydown', handleWindowKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocumentClick)
  window.removeEventListener('keydown', handleWindowKeydown)
})

function handleDocumentClick(event: MouseEvent) {
  if (!root.value?.contains(event.target as Node)) {
    open.value = false
  }
}

function handleWindowKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    open.value = false
  }
}

function toggleOpen() {
  open.value = !open.value
}

function selectOption(value: string) {
  emit('update:modelValue', value)
  emit('change')
  open.value = false
}
</script>

<template>
  <div ref="root" class="fantasy-select" :class="{ 'fantasy-select-stacked': layout === 'stacked' }">
    <input type="hidden" :name="name" :value="modelValue" />
    <span :id="`${id}-label`" class="fantasy-select-label">{{ label }}</span>
    <button
      :id="id"
      type="button"
      class="fantasy-select-button"
      :aria-expanded="open"
      :aria-labelledby="`${id}-label ${id}`"
      aria-haspopup="listbox"
      @click="toggleOpen"
    >
      {{ selectedLabel }}
    </button>
    <ul v-if="open" class="fantasy-select-menu" role="listbox" :aria-labelledby="`${id}-label`">
      <li v-for="option in options" :key="option.value" role="option" :aria-selected="option.value === modelValue">
        <button type="button" class="fantasy-select-option" @click="selectOption(option.value)">
          {{ option.label }}
        </button>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.fantasy-select {
  align-items: center;
  display: flex;
  gap: 0.6rem;
  position: relative;
}

.fantasy-select-stacked {
  align-items: stretch;
  flex-direction: column;
}

.fantasy-select-label {
  background: linear-gradient(
    180deg,
    #ffe9a3,
    #d4af37 45%,
    #8c5a18 75%,
    #4b2805
  );
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  filter: brightness(1.4);
  font-family: 'Cinzel', serif;
  font-size: 1rem;
  letter-spacing: 2px;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 10px rgba(212,175,55,0.75);
}

.fantasy-select-button,
.fantasy-select-option {
  background:
    linear-gradient(rgba(20, 11, 1, 0.88), rgba(20, 11, 1, 0.88)) padding-box,
    linear-gradient(180deg, #ffe9a3, #d4af37 45%, #8c5a18 75%, #4b2805) border-box;
  border: 1px solid transparent;
  border-radius: 0;
  color: #ffe9a3;
  cursor: pointer;
  font-family: 'Cinzel', serif;
  letter-spacing: 1px;
  text-align: left;
  text-shadow: 1px 2px 4px rgba(0,0,0,0.95);
}

.fantasy-select-button {
  min-width: 6.5rem;
  padding: 0.35rem 1.75rem 0.35rem 0.55rem;
}

.fantasy-select-stacked .fantasy-select-button {
  width: 100%;
}

.fantasy-select-button::after {
  content: '⌄';
  position: absolute;
  right: 0.55rem;
}

.fantasy-select-button:focus-visible,
.fantasy-select-option:focus-visible {
  outline: 2px solid rgba(212, 175, 55, 0.85);
  outline-offset: 0.2rem;
}

.fantasy-select-menu {
  background: rgba(20, 11, 1, 0.94);
  border: 1px solid rgba(212, 175, 55, 0.7);
  box-shadow: 0 0.75rem 1.5rem rgba(0, 0, 0, 0.55);
  left: 0;
  list-style: none;
  margin: 0.45rem 0 0;
  min-width: 100%;
  padding: 0.25rem;
  position: absolute;
  top: 100%;
  z-index: 20;
}

.fantasy-select-option {
  background: transparent;
  border: 0;
  display: block;
  padding: 0.45rem 0.55rem;
  width: 100%;
}

.fantasy-select-option:hover,
li[aria-selected='true'] .fantasy-select-option {
  background: rgba(212, 175, 55, 0.16);
  color: #fff0ad;
}
</style>
