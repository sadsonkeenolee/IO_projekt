import './App.css'
import { BrowserRouter, Routes, Route } from "react-router-dom";

import Navbar from "./components/Navbar";
import Header from "./components/Header";

import Home from "./pages/Home";
import Login from "./pages/Login";
import Register from "./pages/Register";
import About from "./pages/About";
import Sources from "./pages/Sources";
import Favorites from "./pages/Favorites";

function App() {
  const token = localStorage.getItem("token");
  const isLoggedIn = !!token;

  return (
    <BrowserRouter>
      <div className="flex flex-col min-h-screen bg-slate-900 text-white">
        
        <header className="flex items-center justify-between bg-slate-800 border-b border-neutral-700 px-6 py-2 shadow-md">
          <Header />
          <Navbar isLoggedIn={isLoggedIn} /> 
        </header>

        <main className="flex-1 p-6 overflow-auto">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/about" element={<About />} />
            <Route path="/sources" element={<Sources />} />
            <Route path="/favorites" element={<Favorites token={token} />} />
          </Routes>
        </main>

      </div>
    </BrowserRouter>
  );
}

export default App;
