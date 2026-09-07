import type { Metadata } from "next";
import "./globals.css";
import { Providers } from "./contexts/Providers";
import Script from "next/script";
import { ErudaDebug } from "../components/ErudaDebug";
// import {NextIntlClientProvider} from 'next-intl';
// import { getLocale } from 'next-intl/server';

export const metadata: Metadata = {
  title: "PPNET",
  description: "PPNET Gaming Platform",
  generator: 'v0.app',
  manifest: "/manifest.webmanifest",
};

type Props = {
  children: React.ReactNode;
};

export default async function RootLayout({
  children
}: Props) {
  // const locale = await getLocale();
  return (
    // Russian is the default application language; LanguageProvider can still
    // switch this value after the user selects another supported language.
    <html lang="ru">
      <body className="antialiased"> 
        <Script src="https://telegram.org/js/telegram-web-app.js" strategy="beforeInteractive" />
        <Script id="telegram-webapp-init" strategy="afterInteractive">
          {`(function(){var w=window.Telegram&&window.Telegram.WebApp;if(w){w.ready();w.expand();}})()`}
        </Script>
        {/* Lightweight in-app debug console (especially useful inside Telegram WebView). */}
        <ErudaDebug />
        <Script id="debug-fallback" strategy="afterInteractive">
          {`(function(){if(document.getElementById('debug-fallback-button'))return;var b=document.createElement('button');b.id='debug-fallback-button';b.textContent='DBG';b.style.cssText='position:fixed;top:12px;right:12px;z-index:2147483647;background:#2563eb;color:#fff;border:2px solid #fff;border-radius:8px;padding:8px 12px;font:bold 12px sans-serif;box-shadow:0 2px 8px #000';b.onclick=function(){var p=document.createElement('pre');p.textContent='Telegram: '+!!window.Telegram+'\\nWebApp: '+!!(window.Telegram&&window.Telegram.WebApp)+'\\ninitData length: '+((window.Telegram&&window.Telegram.WebApp&&window.Telegram.WebApp.initData)||'').length;p.style.cssText='position:fixed;top:60px;right:8px;z-index:2147483647;background:#111;color:#0f0;padding:12px;white-space:pre-wrap';document.body.appendChild(p)};document.body.appendChild(b)})()`}
        </Script>
        {/* <NextIntlClientProvider>{children}</NextIntlClientProvider> */}
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
