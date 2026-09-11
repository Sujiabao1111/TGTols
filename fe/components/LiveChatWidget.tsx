"use client"

import { MessageSquare } from "lucide-react"
import Script from "next/script"
import { openLiveChat } from "@/lib/live-chat"

const liveChatScript = `window.__lc = window.__lc || {};
window.__lc.license = 19930645;
window.__lc.integration_name = "manual_channels";
window.__lc.product_name = "livechat";
window.__lc.asyncInit = true;
;(function(n,t,c){function i(n){return e._h?e._h.apply(null,n):e._q.push(n)}var e={_q:[],_h:null,_v:"2.0",on:function(){i(["on",c.call(arguments)])},once:function(){i(["once",c.call(arguments)])},off:function(){i(["off",c.call(arguments)])},get:function(){if(!e._h)throw new Error("[LiveChatWidget] You can't use getters before load.");return i(["get",c.call(arguments)])},call:function(){i(["call",c.call(arguments)])},init:function(){var n=t.createElement("script");n.async=!0,n.type="text/javascript",n.src="https://cdn.livechatinc.com/tracking.js",t.head.appendChild(n)}};n.LiveChatWidget=n.LiveChatWidget||e}(window,document,[].slice));`

export default function LiveChatWidget() {
  return (
    <>
      <Script id="livechat-widget" strategy="afterInteractive">
        {liveChatScript}
      </Script>

      <button
        type="button"

        onClick={() => openLiveChat()}
        aria-label="Open live chat"
        className="fixed bottom-[4.75rem] right-3 z-50 flex h-14 w-14 items-center justify-center rounded-full bg-[#0b63f6] text-white shadow-[0_10px_30px_rgba(11,99,246,0.45)] transition-transform hover:scale-105 active:scale-95 md:bottom-8 md:right-8"
      >
        <MessageSquare size={28} strokeWidth={2.25} />
      </button>


      <noscript>
        <a href="https://www.livechat.com/chat-with/19930645/" rel="nofollow">
          Chat with us
        </a>
        {", powered by "}
        <a href="https://www.livechat.com/?welcome" rel="noopener nofollow" target="_blank">
          LiveChat
        </a>
      </noscript>
    </>
  )
}


