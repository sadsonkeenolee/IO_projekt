export default function CategorySwitch({ category, setCategory }) {
  const btnBase = "px-6 py-3 rounded-lg font-semibold shadow-md transition";

  return (
    <div className="flex justify-center gap-4 mb-10">

      <button
        onClick={() => setCategory("filmy i seriale")}
        className={`${btnBase} ${
          category === "filmy i seriale"
            ? "bg-emerald-700 ring-4 ring-white"
            : "bg-emerald-600 hover:bg-emerald-700"
        }`}
      >
        📺 Filmy i seriale
      </button>

      <button
        onClick={() => setCategory("ksiazki")}
        className={`${btnBase} ${
          category === "ksiazki"
            ? "bg-indigo-700 ring-4 ring-white"
            : "bg-indigo-600 hover:bg-indigo-700"
        }`}
      >
        📚 Książki
      </button>

    </div>
  );
}