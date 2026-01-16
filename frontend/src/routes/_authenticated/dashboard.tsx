import { getZitadel } from '@/lib/auth'
import { createFileRoute } from '@tanstack/react-router'
import { useEffect, useState } from 'react'

export const Route = createFileRoute('/_authenticated/dashboard')({
    component: RouteComponent,
})

function RouteComponent() {
    const [email, setEmail] = useState('')

    useEffect(() => {
        (async () => {
            const zitadel = getZitadel()
            const user = await zitadel.userManager.getUser()
            setEmail(user?.profile.email ?? '')
        })
    })

    return <div>Welcome {email}</div>
}
