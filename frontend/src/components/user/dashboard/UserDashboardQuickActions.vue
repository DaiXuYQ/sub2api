<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('dashboard.quickActions') }}</h2>
    </div>
    <div class="space-y-3 p-4">
      <button
        v-if="!isSimple"
        type="button"
        :disabled="checkinLoading || checkinStatus?.checked_in_today"
        @click="handleCheckin"
        class="group flex w-full items-center gap-4 rounded-xl border border-emerald-100 bg-emerald-50 p-4 text-left transition-all duration-200 hover:bg-emerald-100 disabled:cursor-not-allowed disabled:opacity-70 dark:border-emerald-900/40 dark:bg-emerald-900/20 dark:hover:bg-emerald-900/30"
      >
        <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-emerald-100 transition-transform group-hover:scale-105 dark:bg-emerald-900/40">
          <Icon name="gift" size="lg" class="text-emerald-600 dark:text-emerald-400" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-gray-900 dark:text-white">
            {{ checkinStatus?.checked_in_today ? '今日已签到' : '每日签到' }}
          </p>
          <p class="text-xs text-gray-500 dark:text-dark-400">
            {{ checkinHint }}
          </p>
        </div>
        <span v-if="checkinLoading" class="h-4 w-4 animate-spin rounded-full border-2 border-emerald-500 border-t-transparent" />
        <Icon
          v-else
          name="chevronRight"
          size="md"
          class="text-emerald-500 transition-colors group-hover:text-emerald-600 dark:text-emerald-400"
        />
      </button>

      <div v-if="checkinMessage" class="rounded-lg bg-emerald-50 px-3 py-2 text-xs text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300">
        {{ checkinMessage }}
      </div>
      <div v-if="checkinError" class="rounded-lg bg-red-50 px-3 py-2 text-xs text-red-600 dark:bg-red-900/20 dark:text-red-300">
        {{ checkinError }}
      </div>

      <button @click="router.push('/keys')" class="group flex w-full items-center gap-4 rounded-xl bg-gray-50 p-4 text-left transition-all duration-200 hover:bg-gray-100 dark:bg-dark-800/50 dark:hover:bg-dark-800">
        <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-primary-100 transition-transform group-hover:scale-105 dark:bg-primary-900/30">
          <Icon name="key" size="lg" class="text-primary-600 dark:text-primary-400" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('dashboard.createApiKey') }}</p>
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('dashboard.generateNewKey') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-gray-400 transition-colors group-hover:text-primary-500 dark:text-dark-500"
        />
      </button>

      <button @click="router.push('/usage')" class="group flex w-full items-center gap-4 rounded-xl bg-gray-50 p-4 text-left transition-all duration-200 hover:bg-gray-100 dark:bg-dark-800/50 dark:hover:bg-dark-800">
        <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-emerald-100 transition-transform group-hover:scale-105 dark:bg-emerald-900/30">
          <Icon name="chart" size="lg" class="text-emerald-600 dark:text-emerald-400" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('dashboard.viewUsage') }}</p>
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('dashboard.checkDetailedLogs') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-gray-400 transition-colors group-hover:text-emerald-500 dark:text-dark-500"
        />
      </button>

      <button @click="router.push('/redeem')" class="group flex w-full items-center gap-4 rounded-xl bg-gray-50 p-4 text-left transition-all duration-200 hover:bg-gray-100 dark:bg-dark-800/50 dark:hover:bg-dark-800">
        <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-amber-100 transition-transform group-hover:scale-105 dark:bg-amber-900/30">
          <Icon name="gift" size="lg" class="text-amber-600 dark:text-amber-400" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('dashboard.redeemCode') }}</p>
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('dashboard.addBalanceWithCode') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-gray-400 transition-colors group-hover:text-amber-500 dark:text-dark-500"
        />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { rewardsAPI, type DailyCheckinStatus } from '@/api'

const props = defineProps<{
  isSimple?: boolean
}>()

const emit = defineEmits<{
  checkinSuccess: []
}>()

const router = useRouter()
const { t } = useI18n()

const checkinStatus = ref<DailyCheckinStatus | null>(null)
const checkinLoading = ref(false)
const checkinMessage = ref('')
const checkinError = ref('')

const checkinHint = computed(() => {
  if (checkinStatus.value?.checked_in_today) {
    const amount = checkinStatus.value.checkin?.reward_amount
    return amount ? `已领取 ${formatAmount(amount)} 余额奖励` : '明天再来领取奖励'
  }
  const min = checkinStatus.value?.reward_min ?? 0.5
  const max = checkinStatus.value?.reward_max ?? 5
  return `随机领取 ${formatAmount(min)}-${formatAmount(max)} 余额`
})

function formatAmount(value: number): string {
  return Number(value || 0).toFixed(2).replace(/\.00$/, '')
}

async function loadCheckinStatus() {
  if (props.isSimple) return
  try {
    checkinStatus.value = await rewardsAPI.getCheckinStatus()
  } catch (error) {
    console.error('Failed to load check-in status:', error)
  }
}

async function handleCheckin() {
  if (checkinLoading.value || checkinStatus.value?.checked_in_today) return
  checkinLoading.value = true
  checkinMessage.value = ''
  checkinError.value = ''
  try {
    const result = await rewardsAPI.checkin()
    checkinStatus.value = {
      checked_in_today: true,
      today: result.checkin.checkin_date,
      reward_min: 0.5,
      reward_max: 5,
      checkin: result.checkin
    }
    checkinMessage.value = `签到成功，获得 ${formatAmount(result.reward_amount)} 余额奖励。当前余额 ${formatAmount(result.balance)}。`
    emit('checkinSuccess')
  } catch (error: any) {
    if (error?.reason === 'DAILY_CHECKIN_ALREADY_DONE') {
      checkinError.value = '今天已经签到过了，明天再来。'
      await loadCheckinStatus()
    } else {
      checkinError.value = error?.message || '签到失败，请稍后重试。'
    }
  } finally {
    checkinLoading.value = false
  }
}

onMounted(() => {
  loadCheckinStatus()
})
</script>
