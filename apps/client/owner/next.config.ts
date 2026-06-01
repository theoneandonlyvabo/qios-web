import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  logging: {
    fetches: {
      fullUrl: false,
      hmrRefreshes: false,
    },
    incomingRequests: false,
  },
};

export default nextConfig;
