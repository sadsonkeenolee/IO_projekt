import { useNavigate } from "react-router-dom";

export default function Logout() {
  const navigate = useNavigate();

  const handleLogout = () => {
    localStorage.removeItem("token");

    navigate("/login");

    window.location.reload();
  };

  return (
    <button
      onClick={handleLogout}
      className="hover:text-red-400 transition-colors duration-200 cursor-pointer flex items-center gap-2"
    >
      🚪 Wyloguj
    </button>
  );
}