import { useState, useEffect } from 'react'
import SiteEditModal from './SiteEditModal'
import { config } from '../config'

const API_BASE = config.apiBaseUrl

function SiteConfigs() {
  const [sites, setSites] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [editingSite, setEditingSite] = useState(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [parserTypeFilter, setParserTypeFilter] = useState('')

  useEffect(() => {
    fetchSites()
  }, [])

  const fetchSites = async () => {
    try {
      setLoading(true)
      const response = await fetch(`${API_BASE}/sites`)
      if (!response.ok) throw new Error('Failed to fetch sites')
      const data = await response.json()
      setSites(data || [])
      setError(null)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = (site) => {
    setEditingSite(site)
    setIsModalOpen(true)
  }

  const handleDelete = async (id) => {
    if (!confirm('Are you sure you want to delete this site?')) return

    try {
      const response = await fetch(`${API_BASE}/sites/${id}`, {
        method: 'DELETE'
      })
      if (!response.ok) throw new Error('Failed to delete site')
      fetchSites()
    } catch (err) {
      alert('Error deleting site: ' + err.message)
    }
  }

  const handleSave = async (site) => {
    try {
      const url = site.id 
        ? `${API_BASE}/sites/${site.id}` 
        : `${API_BASE}/sites`
      const method = site.id ? 'PUT' : 'POST'

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(site)
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Failed to save site')
      }

      setIsModalOpen(false)
      setEditingSite(null)
      fetchSites()
    } catch (err) {
      throw err
    }
  }

  const handleAddNew = () => {
    setEditingSite({
      site_name: '',
      display_name: '',
      base_url: '',
      home_team: '',
      parser_type: 'day_details',
      parser_config: {},
      enabled: true,
      scrape_frequency_hours: 24,
      notes: ''
    })
    setIsModalOpen(true)
  }

  if (loading) {
    return <div className="loading">Loading sites...</div>
  }

  const parserTypes = [...new Set(sites.map(s => s.parser_type))].sort()
  const filteredSites = parserTypeFilter 
    ? sites.filter(s => s.parser_type === parserTypeFilter)
    : sites

  return (
    <div className="sites-container">
      <div className="sites-header">
        <h2>Site Configurations</h2>
        <button className="btn btn-primary" onClick={handleAddNew}>
          + Add New Site
        </button>
      </div>

      {error && <div className="error">{error}</div>}

      <div className="filters">
        <label>
          Parser Type:
          <select 
            value={parserTypeFilter} 
            onChange={(e) => setParserTypeFilter(e.target.value)}
            className="filter-select"
          >
            <option value="">All</option>
            {parserTypes.map(type => (
              <option key={type} value={type}>{type}</option>
            ))}
          </select>
        </label>
      </div>

      <table className="sites-table">
        <thead>
          <tr>
            <th>Site Name</th>
            <th>Display Name</th>
            <th>Base URL</th>
            <th>Parser Type</th>
            <th>Status</th>
            <th>Last Scraped</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {filteredSites.map((site) => (
            <tr key={site.id}>
              <td>{site.site_name}</td>
              <td>{site.display_name}</td>
              <td>
                <a href={site.base_url} target="_blank" rel="noopener noreferrer">
                  {site.base_url}
                </a>
              </td>
              <td>
                <span className="parser-type-badge">{site.parser_type}</span>
              </td>
              <td>
                <span className={`status-badge ${site.enabled ? 'status-enabled' : 'status-disabled'}`}>
                  {site.enabled ? 'Enabled' : 'Disabled'}
                </span>
              </td>
              <td>{site.last_scraped_at || 'Never'}</td>
              <td>
                <div className="action-buttons">
                  <button 
                    className="btn btn-primary btn-small"
                    onClick={() => handleEdit(site)}
                  >
                    Edit
                  </button>
                  <button 
                    className="btn btn-danger btn-small"
                    onClick={() => handleDelete(site.id)}
                  >
                    Delete
                  </button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {isModalOpen && (
        <SiteEditModal
          site={editingSite}
          onSave={handleSave}
          onClose={() => {
            setIsModalOpen(false)
            setEditingSite(null)
          }}
        />
      )}
    </div>
  )
}

export default SiteConfigs
