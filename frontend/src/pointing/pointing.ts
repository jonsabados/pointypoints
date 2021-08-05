import { apiBase } from '@/api/api'
import axios from 'axios'
import { User } from '@/user/user'
import { PointingSession } from '@/pointing/PointingSessionStore'

export interface StartSessionRequest {
  connectionId: string,
  facilitator: User
  facilitatorPoints: boolean
}

export interface Profile {
  email: string
  name: string
  handle: string
}

export async function getProfile(authHeader: string):Promise<Profile> {
  const url = `${apiBase()}/profile`
  const res = await axios.get(url, {
    headers: {
      Authorization: authHeader
    }
  })
  if (res.status !== 200) {
    throw new Error(`unexpected response code ${res.status}`)
  }
  return res.data.result
}

export async function updateProfile(authHeader: string, profile: Profile) {
  const url = `${apiBase()}/profile`
  const res = await axios.put(url, profile, {
    headers: {
      Authorization: authHeader
    }
  })
  if (res.status !== 204) {
    throw new Error(`unexpected response code ${res.status}`)
  }
}

export async function vote(authHeader: string, session: string, userID: string, vote: string) {
  const url = `${apiBase()}/session/${session}/user/${userID}/vote`
  const body = { vote }
  const res = await axios.put(url, body, {
    headers: {
      Authorization: authHeader
    }
  })
  if (res.status !== 204) {
    throw new Error(`unexpected response code ${res.status}`)
  }
}

export async function facilitateSession(authHeader: string, session: string, connectionId: string, facilitatorKey: string): Promise<PointingSession> {
  const url = `${apiBase()}/session/${session}/facilitator`
  const res = await axios.put(url, { connectionId }, {
    headers: {
      Authorization: authHeader,
      'X-Facilitator-Key': facilitatorKey
    }
  })
  if (res.status !== 200) {
    throw new Error(`unexpected response code ${res.status}`)
  }
  return res.data.result
}

export async function joinSession(authHeader: string, session: string, userID: string, user: User) {
  const url = `${apiBase()}/session/${session}/user/${userID}`
  const res = await axios.put(url, user, {
    headers: {
      Authorization: authHeader
    }
  })
  if (res.status !== 204) {
    throw new Error(`unexpected response code ${res.status}`)
  }
  return res.data.result
}

export async function watchSession(authHeader: string, session: string, connectionId: string) {
  const url = `${apiBase()}/session/${session}/watcher`
  const res = await axios.post(url, { connectionId }, {
    headers: {
      Authorization: authHeader
    }
  })
  if (res.status !== 204) {
    throw new Error(`unexpected response code ${res.status}`)
  }
  return res.data.result
}

export async function updateSession(authHeader: string, session: string, facilitatorKey: string, votesShown: boolean, facilitatorPoints: boolean) {
  const url = `${apiBase()}/session/${session}`
  const request = { votesShown, facilitatorPoints }
  const res = await axios.put(url, request, {
    headers: {
      Authorization: authHeader,
      'X-Facilitator-Key': facilitatorKey
    }
  })
  if (res.status !== 204) {
    throw new Error(`unexpected response code ${res.status}`)
  }
}

export async function clearVotes(authHeader: string, session: string, facilitatorKey: string) {
  const url = `${apiBase()}/session/${session}/votes`
  const res = await axios.delete(url, {
    headers: {
      Authorization: authHeader,
      'X-Facilitator-Key': facilitatorKey
    }
  })
  if (res.status !== 204) {
    throw new Error(`unexpected response code ${res.status}`)
  }
}

export async function createSession(authHeader: string, request: StartSessionRequest): Promise<PointingSession> {
  const url = `${apiBase()}/session`
  const res = await axios.post(url, request, {
    headers: {
      Authorization: authHeader
    }
  })
  if (res.status !== 200) {
    throw new Error(`unexpected response code ${res.status}`)
  }
  return res.data.result
}
