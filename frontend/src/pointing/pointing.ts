import { apiBase } from '@/api/api'
import axios from 'axios'

export async function vote(session: string, userID: string, vote: string) {
  const url = `${apiBase()}/session/${session}/user/${userID}/vote`
  const body = { vote }
  const res = await axios.put(url, body, {})
  if (res.status !== 200) {
    throw new Error(`unexpected response code ${res.status}`)
  }
}
