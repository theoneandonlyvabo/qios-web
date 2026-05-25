export default function OwnerDashboardPage() {
  return (
    <main className="min-h-screen bg-[#f4f8ff] px-6 py-8 text-slate-950">
      <section className="mx-auto max-w-5xl rounded-[2rem] bg-white p-8 shadow-[0_24px_80px_rgba(37,99,235,0.12)]">
        <p className="text-sm font-semibold text-[#2f80ed]">Owner Dashboard</p>
        <h1 className="mt-3 text-3xl font-bold tracking-normal">
          Selamat datang di QIOS Owner
        </h1>
        <p className="mt-3 max-w-2xl text-sm leading-6 text-slate-500">
          Login owner berhasil diarahkan ke dashboard owner. Modul dashboard
          penuh dapat dilanjutkan di route ini mengikuti endpoint dashboard QIOS.
        </p>
      </section>
    </main>
  );
}
