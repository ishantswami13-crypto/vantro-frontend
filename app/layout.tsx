import "./globals.css";

export const metadata = {
  title: "VANTRO Dashboard",
  description: "AI-powered finance and billing panel",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="bg-[#05050a] text-white font-sans antialiased">
        {children}
      </body>
    </html>
  );
}
