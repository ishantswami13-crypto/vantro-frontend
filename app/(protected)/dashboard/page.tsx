import Sidebar from "../../components/Sidebar";
import Navbar from "../../components/Navbar";

export default function Dashboard() {
  return (
    <div className="flex min-h-screen">
      <Sidebar />

      <div className="flex flex-col flex-1">
        <Navbar />

        <main className="p-8 space-y-6">
          <h2 className="text-3xl font-bold mb-2">Overview</h2>

          <div className="grid gap-6 grid-cols-1 md:grid-cols-3">
            <Card title="REVENUE (7 DAYS)" value="₹0" />
            <Card title="EXPENSES (7 DAYS)" value="₹0" />
            <Card title="NET (30 DAYS)" value="₹0" />
          </div>
        </main>
      </div>
    </div>
  );
}

function Card({ title, value }: { title: string; value: string }) {
  return (
    <div className="bg-[#111116] rounded-xl border border-[#222] p-5 shadow-sm">
      <p className="text-xs text-gray-400 uppercase tracking-wide">{title}</p>
      <p className="text-2xl font-semibold mt-3">{value}</p>
    </div>
  );
}
