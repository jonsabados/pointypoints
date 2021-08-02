import { apiBase } from '@/api/api'
import axios from 'axios'
import { User } from '@/user/user'
import { PointingSession } from '@/pointing/PointingSessionStore'

export interface StartSessionRequest {
  connectionId: string,
  facilitator: User
  facilitatorPoints: boolean
}

export async function vote(session: string, userID: string, vote: string) {
  const url = `${apiBase()}/session/${session}/user/${userID}/vote`
  const body = { vote }
  const res = await axios.put(url, body, {})
  if (res.status !== 200) {
    throw new Error(`unexpected response code ${res.status}`)
  }
}

export async function joinSession(session: string, userID: string, user: User) {
  const url = `${apiBase()}/session/${session}/user/${userID}`
  const res = await axios.put(url, user, {})
  if (res.status !== 200) {
    throw new Error(`unexpected response code ${res.status}`)
  }
}

export async function watchSession(session: string, connectionId: string) {
  const url = `${apiBase()}/session/${session}/watcher`
  const res = await axios.post(url, { connectionId }, {})
  if (res.status !== 200) {
    throw new Error(`unexpected response code ${res.status}`)
  }
}

export async function updateSession(session: string, facilitatorKey: string, votesShown: boolean, facilitatorPoints: boolean) {
  const url = `${apiBase()}/session/${session}`
  const request = { votesShown, facilitatorPoints }
  const res = await axios.put(url, request, {
    headers: {
      Authorization: facilitatorKey
    }
  })
  if (res.status !== 200) {
    throw new Error(`unexpected response code ${res.status}`)
  }
}

export async function clearVotes(session: string, facilitatorKey: string) {
  const url = `${apiBase()}/session/${session}/votes`
  const res = await axios.delete(url, {
    headers: {
      Authorization: facilitatorKey
    }
  })
  if (res.status !== 200) {
    throw new Error(`unexpected response code ${res.status}`)
  }
}

export async function createSession(request: StartSessionRequest): Promise<PointingSession> {
  const url = `${apiBase()}/session`
  const res = await axios.post(url, request, {})
  if (res.status !== 200) {
    throw new Error(`unexpected response code ${res.status}`)
  }
  return res.data.result
}
