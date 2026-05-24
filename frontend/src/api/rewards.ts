/**
 * Daily reward/check-in API endpoints
 */

import { apiClient } from './client'

export interface DailyCheckin {
  id: number
  user_id: number
  checkin_date: string
  reward_amount: number
  redeem_code_id?: number
  created_at: string
}

export interface DailyCheckinStatus {
  checked_in_today: boolean
  today: string
  reward_min: number
  reward_max: number
  checkin?: DailyCheckin
}

export interface DailyCheckinResult {
  checked_in_today: boolean
  reward_amount: number
  balance: number
  checkin: DailyCheckin
}

export async function getCheckinStatus(): Promise<DailyCheckinStatus> {
  const { data } = await apiClient.get<DailyCheckinStatus>('/rewards/checkin/status')
  return data
}

export async function checkin(): Promise<DailyCheckinResult> {
  const { data } = await apiClient.post<DailyCheckinResult>('/rewards/checkin')
  return data
}

export const rewardsAPI = {
  getCheckinStatus,
  checkin
}

export default rewardsAPI
