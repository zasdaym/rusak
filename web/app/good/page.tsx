import { captureException } from "@sentry/react"

async function getData() {
  try {
    const response = await fetch(`${process.env.API_URL}/good`, { cache: 'no-store' })
    if (!response.ok) {
      throw new Error(response.statusText)
    }
    const data = await response.text()
    return data
  } catch (e) {
    captureException(e, {
      user: {
        id: 666,
        email: 'zasdaym@gmail.com' // Pretend this is extracted from the session,
      }
    })
  }
}

export default async function Page() {
  const data = await getData() || ""

  return (
    <>
      <h1>Good</h1>
      <p>{data}</p>
    </>
  )
}
