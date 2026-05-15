import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'FC Centrum Spots Map',
  description: 'Visualize FC Centrum spots on an interactive Amsterdam map.',
  icons: {
    icon: 'https://fccentrum.nl/favicon.ico',
    shortcut: 'https://fccentrum.nl/favicon.ico',
    apple: 'https://fccentrum.nl/favicon.ico',
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
