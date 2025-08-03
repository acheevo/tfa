
import { SocialIcon } from '../ui/SocialIcons';

export default function Footer() {
  const footerLinks = {
    'Game': ['Features', 'Classes', 'Races', 'World Map'],
    'Community': ['Forums', 'Discord', 'Reddit', 'Wiki'],
    'Support': ['Help Center', 'Contact Us', 'Bug Report', 'FAQ'],
    'Legal': ['Terms of Service', 'Privacy Policy', 'Cookie Policy', 'EULA']
  };

  const socialLinks = [
    { platform: 'youtube' as const, href: 'https://youtube.com/@thefoolsascension' },
    { platform: 'discord' as const, href: 'https://discord.gg/thefoolsascension' },
    { platform: 'instagram' as const, href: 'https://instagram.com/thefoolsascension' },
    { platform: 'x' as const, href: 'https://x.com/thefoolsascension' },
    { platform: 'reddit' as const, href: 'https://reddit.com/r/thefoolsascension' }
  ];

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
            <p className="text-gray-600 text-sm font-cormorant">
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
                      className="text-gray-600 hover:text-teal-700 text-sm transition-colors duration-200 font-cormorant"
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
            <p className="text-gray-600 text-sm mb-4 md:mb-0 font-cormorant">
              © 2025 The Fool's Ascension. All rights reserved.
            </p>
            
            {/* Social Links */}
            <div className="flex space-x-6">
              {socialLinks.map((social) => (
                <SocialIcon
                  key={social.platform}
                  platform={social.platform}
                  href={social.href}
                />
              ))}
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
}