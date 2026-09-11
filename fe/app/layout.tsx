import type { Metadata } from "next";
import "./globals.css";
import { Providers } from "./contexts/Providers";
import Script from "next/script";
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
    <html lang="en">
      <body className="antialiased"> 
        <Script src="https://telegram.org/js/telegram-web-app.js" strategy="beforeInteractive" />
        <Script id="telegram-webapp-init" strategy="afterInteractive">
          {`(function(){var w=window.Telegram&&window.Telegram.WebApp;if(w){w.ready();w.expand();}})()`}
        </Script>
        {/* <NextIntlClientProvider>{children}</NextIntlClientProvider> */}
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
