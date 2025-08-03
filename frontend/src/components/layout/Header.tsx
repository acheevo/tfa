
export default function Header() {
  const menuItems = ['HOME', 'PAGES', 'FORUM', 'PORTFOLIO', 'BLOG', 'SHOP', 'ELEMENTS'];

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
                    className="text-gray-700 text-xs md:text-sm font-uncial hover:text-gray-900 transition-colors duration-300 tracking-wider"
                  >
                    {item}
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