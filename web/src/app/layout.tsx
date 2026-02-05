import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { Providers } from "./providers";
import { AppDynamicsRUM } from "@/components/AppDynamicsRUM";
import { getRUMConfig } from "@/lib/rum-config";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Claudeception - Rule Management",
  description: "Manage Claude Code rules across teams and projects",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const rumConfig = getRUMConfig();

  return (
    <html lang="en">
      <head>
        {rumConfig.enabled && (
          <AppDynamicsRUM
            appKey={rumConfig.appKey}
            adrumExtUrl={rumConfig.adrumExtUrl}
            beaconUrl={rumConfig.beaconUrl}
          />
        )}
      </head>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
