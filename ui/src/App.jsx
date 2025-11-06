import { useState } from 'react'
import './App.css'
import SiteConfigs from './components/SiteConfigs'

function App() {
  const [activeView, setActiveView] = useState('home')

  return (
    <div className="app">
      <nav className="navbar">
        <div className="navbar-brand">
          <h1>Hockey Calendar Scraper</h1>
        </div>
        <div className="navbar-menu">
          <a 
            href="#" 
            className={activeView === 'home' ? 'active' : ''}
            onClick={() => setActiveView('home')}
          >
            Home
          </a>
          <a 
            href="#" 
            className={activeView === 'sites' ? 'active' : ''}
            onClick={() => setActiveView('sites')}
          >
            Site Configs
          </a>
        </div>
      </nav>

      <main className="main-content">
        {activeView === 'home' && (
          <div className="home">
            <h2>Welcome to Hockey Calendar Scraper</h2>
            <p>Manage site configurations and scrape hockey schedules.</p>
          </div>
        )}
        
        {activeView === 'sites' && <SiteConfigs />}
      </main>
    </div>
  )
}

export default App
