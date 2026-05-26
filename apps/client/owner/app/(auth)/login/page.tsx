// Login page dihandle oleh Dev 3 di apps/client/web/
// File ini adalah placeholder agar route /login tidak 404
// Redirect ke auth app jika diakses langsung

import { redirect } from 'next/navigation';

export default function LoginPage() {
  // Jika ada NEXT_PUBLIC_AUTH_URL, redirect ke sana
  // Untuk sekarang redirect ke splash
  redirect('/');
}
