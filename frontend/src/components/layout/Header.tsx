
export default function Header() {
  const menuItems = ['Home', 'Pages', 'Forum', 'Portfolio', 'Blog', 'Shop', 'Elements'];

  return (
    <header className="absolute top-0 w-full z-50 py-4 md:py-8">
      <div className="container mx-auto px-4">
        <div className="flex flex-col items-center">
          <div className="mb-4 md:mb-6">
            <img 
              src="/src/assets/img/logo.png" 
              alt="The Fool's Ascension" 
              className="h-24 md:h-32 lg:h-36 w-auto"
            />
          </div>
          <nav>
            <ul className="flex flex-wrap justify-center gap-4 md:gap-6 lg:gap-8">
              {menuItems.map((item) => (
                <li key={item}>
                  <a 
                    href="#" 
                    className="text-gray-700 text-sm md:text-base font-uncial hover:text-gray-900 transition-colors duration-300 tracking-wider relative group"
                  >
                    {item}
                    <span className="absolute -bottom-1 left-0 w-0 h-0.5 bg-gradient-to-r from-teal-600 via-cyan-700 to-teal-800 transition-all duration-500 ease-out group-hover:w-[35%]"></span>
                  </a>
                </li>
              ))}
            </ul>
          </nav>
        </div>
      </div>
    </header>
  );
}