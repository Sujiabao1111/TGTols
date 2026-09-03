/** @type {import('next').NextConfig} */
const nextConfig = {
  async headers() {
    return [{
      source: '/tonconnect-manifest.json',
      headers: [
        { key: 'Access-Control-Allow-Origin', value: '*' },
        { key: 'Content-Type', value: 'application/manifest+json' },
      ],
    }]
  },
  typescript: {
    ignoreBuildErrors: true,
  },
  images: {
    unoptimized: true,
  },
}

export default nextConfig
