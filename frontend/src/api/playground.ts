import { apiClient } from './client'
import type { PaginatedResponse } from '@/types'

export interface PlaygroundChatSessionRecord {
  id: number
  user_id: number
  title: string
  model: string
  api_key_id?: number | null
  system_prompt: string
  use_context: boolean
  metadata?: Record<string, unknown>
  last_message_at: string
  created_at: string
  updated_at: string
}

export interface PlaygroundChatMessageRecord {
  id: number
  session_id: number
  user_id: number
  api_key_id?: number | null
  role: 'user' | 'assistant'
  model: string
  content: string
  content_json?: Record<string, unknown>
  images?: Array<Record<string, unknown>>
  usage?: Record<string, unknown>
  status: string
  error: string
  duration_ms?: number | null
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface PlaygroundImageTaskRecord {
  id: number
  user_id: number
  api_key_id?: number | null
  model: string
  prompt: string
  quality: string
  size: string
  n: number
  endpoint: string
  status: 'pending' | 'running' | 'success' | 'error'
  request?: Record<string, unknown>
  reference_images?: Array<Record<string, unknown>>
  result_images?: Array<Record<string, unknown>>
  response?: Record<string, unknown>
  error: string
  usage?: Record<string, unknown>
  cost: number
  duration_ms?: number | null
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface CreateChatSessionPayload {
  title?: string
  model?: string
  api_key_id?: number | null
  system_prompt?: string
  use_context?: boolean
  metadata?: Record<string, unknown>
}

export interface CreateChatMessagePayload {
  api_key_id?: number | null
  role: 'user' | 'assistant'
  model?: string
  content?: string
  content_json?: Record<string, unknown>
  images?: Array<Record<string, unknown>>
  usage?: Record<string, unknown>
  status?: string
  error?: string
  duration_ms?: number | null
  metadata?: Record<string, unknown>
}

export interface CreateImageTaskPayload {
  api_key_id?: number | null
  model: string
  prompt: string
  quality: string
  size: string
  n: number
  endpoint?: string
  status?: 'pending' | 'running' | 'success' | 'error'
  request?: Record<string, unknown>
  reference_images?: Array<Record<string, unknown>>
  result_images?: Array<Record<string, unknown>>
  response?: Record<string, unknown>
  error?: string
  usage?: Record<string, unknown>
  cost?: number
  duration_ms?: number | null
  metadata?: Record<string, unknown>
}

export type UpdateImageTaskPayload = Partial<Pick<CreateImageTaskPayload, 'status' | 'result_images' | 'response' | 'error' | 'usage' | 'cost' | 'duration_ms' | 'metadata'>>

export async function listChatSessions(page = 1, pageSize = 30): Promise<PaginatedResponse<PlaygroundChatSessionRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<PlaygroundChatSessionRecord>>('/playground/chat/sessions', { params: { page, page_size: pageSize } })
  return data
}

export async function createChatSession(payload: CreateChatSessionPayload): Promise<PlaygroundChatSessionRecord> {
  const { data } = await apiClient.post<PlaygroundChatSessionRecord>('/playground/chat/sessions', payload)
  return data
}

export async function updateChatSession(id: number, payload: CreateChatSessionPayload): Promise<PlaygroundChatSessionRecord> {
  const { data } = await apiClient.put<PlaygroundChatSessionRecord>(`/playground/chat/sessions/${id}`, payload)
  return data
}

export async function deleteChatSession(id: number): Promise<void> {
  await apiClient.delete(`/playground/chat/sessions/${id}`)
}

export async function listChatMessages(sessionId: number, page = 1, pageSize = 200): Promise<PaginatedResponse<PlaygroundChatMessageRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<PlaygroundChatMessageRecord>>(`/playground/chat/sessions/${sessionId}/messages`, { params: { page, page_size: pageSize } })
  return data
}

export async function createChatMessage(sessionId: number, payload: CreateChatMessagePayload): Promise<PlaygroundChatMessageRecord> {
  const { data } = await apiClient.post<PlaygroundChatMessageRecord>(`/playground/chat/sessions/${sessionId}/messages`, payload)
  return data
}

export async function listImageTasks(page = 1, pageSize = 20): Promise<PaginatedResponse<PlaygroundImageTaskRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<PlaygroundImageTaskRecord>>('/playground/image/tasks', { params: { page, page_size: pageSize } })
  return data
}

export async function createImageTask(payload: CreateImageTaskPayload): Promise<PlaygroundImageTaskRecord> {
  const { data } = await apiClient.post<PlaygroundImageTaskRecord>('/playground/image/tasks', payload)
  return data
}

export async function updateImageTaskRemote(id: number, payload: UpdateImageTaskPayload): Promise<PlaygroundImageTaskRecord> {
  const { data } = await apiClient.put<PlaygroundImageTaskRecord>(`/playground/image/tasks/${id}`, payload)
  return data
}

export async function deleteImageTask(id: number): Promise<void> {
  await apiClient.delete(`/playground/image/tasks/${id}`)
}

export const playgroundAPI = {
  listChatSessions,
  createChatSession,
  updateChatSession,
  deleteChatSession,
  listChatMessages,
  createChatMessage,
  listImageTasks,
  createImageTask,
  updateImageTask: updateImageTaskRemote,
  deleteImageTask,
}

export default playgroundAPI
