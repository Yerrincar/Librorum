<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

type CalendarDay = {
  iso: string
  label: string
  inMonth: boolean
  selected: boolean
  today: boolean
}

const props = defineProps<{
  id: string
  label: string
  modelValue: string
  name: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  change: []
}>()

const open = ref(false)
const root = ref<HTMLElement>()
const visibleMonth = ref(startOfMonth(new Date()))
const hour = ref('00')
const minute = ref('00')

const weekdays = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']

const monthFormatter = new Intl.DateTimeFormat(undefined, { month: 'long', year: 'numeric' })
const displayFormatter = new Intl.DateTimeFormat(undefined, {
  day: '2-digit',
  month: 'short',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
})

const selectedDate = computed(() => parseDateTime(props.modelValue))
const selectedDatePart = computed(() => (selectedDate.value ? formatDatePart(selectedDate.value) : ''))
const monthLabel = computed(() => monthFormatter.format(visibleMonth.value))
const displayValue = computed(() => (selectedDate.value ? displayFormatter.format(selectedDate.value) : 'Choose date'))

const calendarDays = computed<CalendarDay[]>(() => {
  const firstDayOfMonth = startOfMonth(visibleMonth.value)
  const firstWeekday = (firstDayOfMonth.getDay() + 6) % 7
  const startDate = new Date(firstDayOfMonth)
  startDate.setDate(firstDayOfMonth.getDate() - firstWeekday)

  return Array.from({ length: 42 }, (_, index) => {
    const day = new Date(startDate)
    day.setDate(startDate.getDate() + index)
    const iso = formatDatePart(day)

    return {
      iso,
      label: String(day.getDate()),
      inMonth: day.getMonth() === visibleMonth.value.getMonth(),
      selected: iso === selectedDatePart.value,
      today: iso === formatDatePart(new Date()),
    }
  })
})

watch(
  () => props.modelValue,
  (value) => {
    const date = parseDateTime(value)
    if (!date) {
      return
    }

    visibleMonth.value = startOfMonth(date)
    hour.value = pad(date.getHours())
    minute.value = pad(date.getMinutes())
  },
  { immediate: true },
)

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
  if (open.value && selectedDate.value) {
    visibleMonth.value = startOfMonth(selectedDate.value)
  }
}

function previousMonth() {
  visibleMonth.value = addMonths(visibleMonth.value, -1)
}

function nextMonth() {
  visibleMonth.value = addMonths(visibleMonth.value, 1)
}

function selectDay(iso: string) {
  updateValue(`${iso}T${normalizedHour()}:${normalizedMinute()}`)
}

function selectNow() {
  const now = new Date()
  hour.value = pad(now.getHours())
  minute.value = pad(now.getMinutes())
  visibleMonth.value = startOfMonth(now)
  updateValue(formatDateTime(now))
}

function clearValue() {
  updateValue('')
  open.value = false
}

function updateHour(event: Event) {
  hour.value = timeInputValue(event)
}

function updateMinute(event: Event) {
  minute.value = timeInputValue(event)
}

function commitTime() {
  hour.value = normalizedHour()
  minute.value = normalizedMinute()

  if (selectedDatePart.value) {
    updateValue(`${selectedDatePart.value}T${hour.value}:${minute.value}`)
  }
}

function updateValue(value: string) {
  emit('update:modelValue', value)
  emit('change')
}

function timeInputValue(event: Event) {
  return ((event.target as HTMLInputElement).value.match(/\d/g) ?? []).join('').slice(0, 2)
}

function normalizedHour() {
  return normalizeTimePart(hour.value, 23)
}

function normalizedMinute() {
  return normalizeTimePart(minute.value, 59)
}

function normalizeTimePart(value: string, max: number) {
  const parsed = Number(value)
  if (!Number.isFinite(parsed)) {
    return '00'
  }
  return pad(Math.min(Math.max(parsed, 0), max))
}

function parseDateTime(value: string) {
  const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})$/.exec(value)
  if (!match) {
    return null
  }

  const [, year, month, day, hours, minutes] = match
  const date = new Date(Number(year), Number(month) - 1, Number(day), Number(hours), Number(minutes))
  if (Number.isNaN(date.getTime())) {
    return null
  }

  return date
}

function startOfMonth(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), 1)
}

function addMonths(date: Date, amount: number) {
  return new Date(date.getFullYear(), date.getMonth() + amount, 1)
}

function formatDateTime(date: Date) {
  return `${formatDatePart(date)}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function formatDatePart(date: Date) {
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

function pad(value: number) {
  return String(value).padStart(2, '0')
}
</script>

<template>
  <div ref="root" class="fantasy-date-time">
    <input type="hidden" :name="name" :value="modelValue" />
    <span :id="`${id}-label`" class="fantasy-date-time-label">{{ label }}</span>
    <button
      :id="id"
      type="button"
      class="fantasy-date-time-button"
      :aria-expanded="open"
      :aria-labelledby="`${id}-label ${id}`"
      aria-haspopup="dialog"
      @click="toggleOpen"
    >
      {{ displayValue }}
    </button>

    <section v-if="open" class="fantasy-date-time-popover" role="dialog" :aria-labelledby="`${id}-label`">
      <header class="calendar-header">
        <button type="button" aria-label="Previous month" @click="previousMonth">‹</button>
        <span>{{ monthLabel }}</span>
        <button type="button" aria-label="Next month" @click="nextMonth">›</button>
      </header>

      <div class="calendar-weekdays" aria-hidden="true">
        <span v-for="weekday in weekdays" :key="weekday">{{ weekday }}</span>
      </div>

      <div class="calendar-days">
        <button
          v-for="day in calendarDays"
          :key="day.iso"
          type="button"
          class="calendar-day"
          :class="{
            'calendar-day-muted': !day.inMonth,
            'calendar-day-selected': day.selected,
            'calendar-day-today': day.today,
          }"
          @click="selectDay(day.iso)"
        >
          {{ day.label }}
        </button>
      </div>

      <div class="time-row">
        <label>
          Hour
          <input :value="hour" inputmode="numeric" maxlength="2" @blur="commitTime" @input="updateHour" />
        </label>
        <label>
          Minute
          <input :value="minute" inputmode="numeric" maxlength="2" @blur="commitTime" @input="updateMinute" />
        </label>
      </div>

      <footer class="calendar-actions">
        <button type="button" @click="selectNow">Now</button>
        <button type="button" @click="clearValue">Clear</button>
      </footer>
    </section>
  </div>
</template>

<style scoped>
.fantasy-date-time {
  display: grid;
  gap: 0.35rem;
  min-width: 0;
  position: relative;
}

.fantasy-date-time-label {
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

.fantasy-date-time-button,
.fantasy-date-time-popover button,
.time-row input {
  background:
    linear-gradient(rgba(20, 11, 1, 0.88), rgba(20, 11, 1, 0.88)) padding-box,
    linear-gradient(180deg, #ffe9a3, #d4af37 45%, #8c5a18 75%, #4b2805) border-box;
  border: 1px solid transparent;
  border-radius: 0;
  color: #ffe9a3;
  font-family: 'Cinzel', serif;
  letter-spacing: 1px;
  text-shadow: 1px 2px 4px rgba(0,0,0,0.95);
}

.fantasy-date-time-button {
  cursor: pointer;
  padding: 0.35rem 1.75rem 0.35rem 0.55rem;
  position: relative;
  text-align: left;
  width: 100%;
}

.fantasy-date-time-button::after {
  content: '⌄';
  position: absolute;
  right: 0.55rem;
}

.fantasy-date-time-button:focus-visible,
.fantasy-date-time-popover button:focus-visible,
.time-row input:focus-visible {
  outline: 2px solid rgba(212, 175, 55, 0.85);
  outline-offset: 0.2rem;
}

.fantasy-date-time-popover {
  background: rgba(20, 11, 1, 0.96);
  border: 1px solid rgba(212, 175, 55, 0.72);
  box-shadow: 0 0.75rem 1.5rem rgba(0, 0, 0, 0.55);
  box-sizing: border-box;
  display: grid;
  gap: 0.65rem;
  left: 0;
  padding: 0.75rem;
  position: absolute;
  top: calc(100% + 0.45rem);
  width: min(21rem, calc(100vw - 3rem));
  z-index: 60;
}

.calendar-header,
.calendar-actions {
  align-items: center;
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
}

.calendar-header span {
  background: linear-gradient(180deg, #ffe9a3, #d4af37 45%, #8c5a18 75%, #4b2805);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  filter: brightness(1.35);
  font-family: 'Cinzel', serif;
  letter-spacing: 2px;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 10px rgba(212,175,55,0.75);
}

.fantasy-date-time-popover button {
  cursor: pointer;
  padding: 0.35rem 0.5rem;
}

.calendar-weekdays,
.calendar-days {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 0.25rem;
}

.calendar-weekdays span {
  color: #cfa34a;
  font-family: 'Cinzel', serif;
  font-size: 0.72rem;
  letter-spacing: 1px;
  text-align: center;
  text-shadow: 1px 2px 4px rgba(0,0,0,0.9);
}

.calendar-day {
  aspect-ratio: 1;
  min-width: 0;
}

.calendar-day-muted {
  opacity: 0.42;
}

.calendar-day-today {
  box-shadow: inset 0 0 0 1px rgba(255, 233, 163, 0.55);
}

.calendar-day-selected {
  background:
    linear-gradient(rgba(90, 53, 8, 0.84), rgba(90, 53, 8, 0.84)) padding-box,
    linear-gradient(180deg, #fff0ad, #d4af37 55%, #8c5a18) border-box;
  color: #fff0ad;
}

.time-row {
  display: grid;
  gap: 0.5rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.time-row label {
  color: #cfa34a;
  display: grid;
  font-family: 'Cinzel', serif;
  font-size: 0.8rem;
  gap: 0.25rem;
  letter-spacing: 1px;
  text-shadow: 1px 2px 4px rgba(0,0,0,0.9);
}

.time-row input {
  box-sizing: border-box;
  padding: 0.35rem 0.45rem;
  width: 100%;
}
</style>
