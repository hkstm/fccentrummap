import type { NextConfig } from 'next';

const rawBasePath = process.env.PAGES_BASE_PATH?.trim() ?? '';
const normalizedBasePath = rawBasePath.replace(/^\/+|\/+$/g, '');
const basePath = normalizedBasePath ? `/${normalizedBasePath}` : '';

const assetPrefix = process.env.CDN_ASSET_PREFIX?.trim() || undefined;

const nextConfig: NextConfig = {
  output: 'export',
  trailingSlash: true,
  images: { unoptimized: true },
  basePath,
  assetPrefix,
  env: {
    NEXT_PUBLIC_BASE_PATH: basePath,
  },
};

export default nextConfig;
