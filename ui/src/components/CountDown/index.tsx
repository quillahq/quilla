import { useEffect, useState } from "react"

const fixedZero = (time: number) => time * 1 < 10 ? `0${time}` : time

export default function CountDown(props: { date: Date }) {
    const getTime = () => props.date.getTime() - new Date().getTime()

    const [time, setTime] = useState(getTime())

    useEffect(() => {
        const interval = setInterval(() => setTime(getTime()), 1000)
        return () => {
            clearInterval(interval)
        }
    })

    const hours = 60 * 60 * 1000
    const minutes = 60 * 1000

    const h = Math.floor(time/hours)
    const m = Math.floor((time - h * hours) / minutes)
    const s = Math.floor((time - h * hours - m * minutes) / 1000)
    return <>{fixedZero(h)}:{fixedZero(m)}:{fixedZero(s)}</>
}