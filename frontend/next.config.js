/** @type {import('next').NextConfig} */
const nextConfig = {
  // Amplify suporta SSR automaticamente
  experimental: {
    serverComponentsExternalPackages: [],
  },

  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
  },

  // Otimizações para Amplify
  swcMinify: true,

  // Headers de segurança
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
        ],
      },
    ]
  },
}

module.exports = nextConfig
