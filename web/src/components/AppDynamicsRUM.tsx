'use client';

import Script from 'next/script';

interface AppDynamicsRUMProps {
  appKey: string;
  adrumExtUrl: string;
  beaconUrl: string;
}

declare global {
  interface Window {
    'adrum-start-time': number;
    ADRUM: {
      config: {
        appKey: string;
        adrumExtUrlHttp: string;
        adrumExtUrlHttps: string;
        beaconUrlHttp: string;
        beaconUrlHttps: string;
        xd: { enable: boolean };
        spa: { spa2: boolean };
        resTiming: { bufSize: number; clearResTimingOnBeaconSend: boolean };
        maxUrlLength: number;
      };
    };
  }
}

export function AppDynamicsRUM({ appKey, adrumExtUrl, beaconUrl }: AppDynamicsRUMProps) {
  if (!appKey || !adrumExtUrl || !beaconUrl) {
    return null;
  }

  const initScript = `
    window['adrum-start-time'] = new Date().getTime();
    window.ADRUM = window.ADRUM || {};
    window.ADRUM.config = {
      appKey: '${appKey}',
      adrumExtUrlHttp: '${adrumExtUrl}',
      adrumExtUrlHttps: '${adrumExtUrl}',
      beaconUrlHttp: '${beaconUrl}',
      beaconUrlHttps: '${beaconUrl}',
      xd: { enable: false },
      spa: { spa2: true },
      resTiming: { bufSize: 200, clearResTimingOnBeaconSend: true },
      maxUrlLength: 512
    };
  `;

  return (
    <>
      <Script
        id="adrum-config"
        strategy="beforeInteractive"
        dangerouslySetInnerHTML={{ __html: initScript }}
      />
      <Script
        id="adrum-agent"
        strategy="beforeInteractive"
        src={`${adrumExtUrl}/adrum-ext.0.js`}
      />
    </>
  );
}
