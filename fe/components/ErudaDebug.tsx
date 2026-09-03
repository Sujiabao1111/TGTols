"use client"

import { useEffect, useState } from "react"

export function ErudaDebug() {
  const [open, setOpen] = useState(false)
  const [logs, setLogs] = useState<string[]>([])
  useEffect(() => {
    const original = console.info
    const capture = (...args: unknown[]) => { setLogs((prev) => [...prev.slice(-49), args.map(String).join(" ")]); original(...args) }
    console.info = capture
    const existing = document.querySelector('script[data-eruda-debug="true"]')
    if (existing) return () => { console.info = original }

    const script = document.createElement("script")
    // Use the bundled copy so Telegram WebView/CDN blocking cannot prevent debugging.
    script.src = "/eruda.js"
    script.async = true
    script.dataset.erudaDebug = "true"
    const initEruda = () => {
      const eruda = (window as typeof window & { eruda?: { init: () => void } }).eruda
      if (eruda) {
        eruda.init()
        console.info("[debug] Eruda console initialized")
      } else {
        console.warn("[debug] Eruda loaded but global was not found")
      }
    }
    script.onload = initEruda
    script.onerror = () => console.error("[debug] Local /eruda.js failed to load")
    document.head.appendChild(script)
    window.addEventListener("load", initEruda, { once: true })
    return () => { console.info = original; window.removeEventListener("load", initEruda) }
  }, [])

  return <>
    <button type="button" onClick={() => setOpen(!open)} style={{position:"fixed",right:16,top:16,zIndex:2147483647,height:36,padding:"0 10px",borderRadius:8,background:"#2563eb",color:"white",fontSize:12,fontWeight:700,border:"2px solid white"}}>DBG</button>
    {open && <div style={{position:"fixed",right:8,top:60,zIndex:2147483647,width:320,maxHeight:360,overflow:"auto",background:"#111827",color:"#d1d5db",padding:12,borderRadius:8,fontSize:11,whiteSpace:"pre-wrap"}}>
      <b style={{color:"white"}}>Debug console</b>{logs.length ? logs.map((log, i) => <div key={i}>{log}</div>) : <div>No logs yet</div>}
    </div>}
  </>
}
