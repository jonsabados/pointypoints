import { apiBase } from '@/api/api'
import axios from 'axios'
import { User } from '@/user/user'

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
