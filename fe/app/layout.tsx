import type { Metadata } from "next";
import "./globals.css";
import { Providers } from "./contexts/Providers";
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
    // <html lang={locale}>
    <html lang="en">
      <body className="antialiased"> 
        {/* <NextIntlClientProvider>{children}</NextIntlClientProvider> */}
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
