
export default function Footer() {
  const footerLinks = {
    'Game': ['Features', 'Classes', 'Races', 'World Map'],
    'Community': ['Forums', 'Discord', 'Reddit', 'Wiki'],
    'Support': ['Help Center', 'Contact Us', 'Bug Report', 'FAQ'],
    'Legal': ['Terms of Service', 'Privacy Policy', 'Cookie Policy', 'EULA']
  };

  return (
    <footer className="bg-gradient-to-b from-slate-50 to-gray-100 text-gray-800 py-16 border-t border-teal-200">
      <div className="container mx-auto px-4">
        <div className="grid grid-cols-1 md:grid-cols-5 gap-8 mb-12">
          {/* Logo and Description */}
          <div className="md:col-span-1">
            <img 
              src="/src/assets/img/logo.png" 
              alt="The Fool's Ascension" 
              className="h-20 w-auto mb-4"
            />
            <p className="text-gray-600 text-sm font-cinzel">
              Embark on an epic journey in a world of magic and mystery.
            </p>
          </div>

          {/* Footer Links */}
          {Object.entries(footerLinks).map(([category, links]) => (
            <div key={category}>
              <h3 className="font-uncial text-lg mb-4 bg-gradient-to-r from-teal-600 to-cyan-700 bg-clip-text text-transparent">{category}</h3>
              <ul className="space-y-2">
                {links.map((link) => (
                  <li key={link}>
                    <a 
                      href="#" 
                      className="text-gray-600 hover:text-teal-700 text-sm transition-colors duration-200 font-cinzel"
                    >
                      {link}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Divider */}
        <div className="border-t border-teal-200 pt-8">
          <div className="flex flex-col md:flex-row justify-between items-center">
            <p className="text-gray-600 text-sm mb-4 md:mb-0 font-cinzel">
              © 2025 The Fool's Ascension. All rights reserved.
            </p>
            
            {/* Social Links */}
            <div className="flex space-x-6">
              {['Twitter', 'Facebook', 'Instagram', 'YouTube'].map((social) => (
                <a 
                  key={social}
                  href="#" 
                  className="text-gray-500 hover:text-teal-600 transition-colors duration-200"
                >
                  <span className="sr-only">{social}</span>
                  <div className="w-6 h-6 bg-gradient-to-r from-teal-200 to-cyan-200 rounded-full hover:from-teal-300 hover:to-cyan-300 transition-all duration-200"></div>
                </a>
              ))}
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
}