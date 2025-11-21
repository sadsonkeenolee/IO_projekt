export default function Suggestions() {
  return (
    <div>
      <h2 className="text-2xl font-bold mb-6 text-center">Twoje podpowiedzi</h2>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">

        <div className="bg-slate-700 p-5 rounded-xl shadow-md hover:shadow-lg border border-slate-600 transition">
          <h3 className="text-lg font-semibold mb-2">🎬 Przykładowy tytuł</h3>
          <p className="text-sm text-slate-300">Opis pozycji...</p>
        </div>

        <div className="bg-slate-700 p-5 rounded-xl shadow-md hover:shadow-lg border border-slate-600 transition">
          <h3 className="text-lg font-semibold mb-2">📚 Przykładowa książka</h3>
          <p className="text-sm text-slate-300">Opis pozycji...</p>
        </div>

        <div className="bg-slate-700 p-5 rounded-xl shadow-md hover:shadow-lg border border-slate-600 transition">
          <h3 className="text-lg font-semibold mb-2">📺 Przykładowy serial</h3>
          <p className="text-sm text-slate-300">Opis pozycji...</p>
        </div>

      </div>
    </div>
  );
}
