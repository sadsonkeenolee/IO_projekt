import { useState } from "react";
import CategorySwitch from "../components/CategorySwitch";
import MainPanel from "../components/MainPanel";
import Suggestions from "../components/Suggestions";

export default function Home() {
  const [category, setCategory] = useState("ksiazki");

  return (
    <main className={`flex-1 p-10 transition-colors duration-500 bg-slate-900`}>
      <CategorySwitch category={category} setCategory={setCategory} />
      <MainPanel category={category} />
      <Suggestions />
    </main>
  );
}
