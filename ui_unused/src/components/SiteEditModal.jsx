import { useState, useEffect } from 'react'

function SiteEditModal({ site, onSave, onClose }) {
  const [formData, setFormData] = useState(site)
  const [parserConfigText, setParserConfigText] = useState('')
  const [error, setError] = useState(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    // Convert parser_config object to pretty JSON string
    if (site.parser_config) {
      setParserConfigText(JSON.stringify(site.parser_config, null, 2))
    }
  }, [site])

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }))
  }

  const handleParserConfigChange = (e) => {
    setParserConfigText(e.target.value)
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError(null)
    setSaving(true)

    try {
      // Parse parser_config JSON
      let parserConfig = {}
      if (parserConfigText.trim()) {
        try {
          parserConfig = JSON.parse(parserConfigText)
        } catch (err) {
          throw new Error('Invalid JSON in Parser Config: ' + err.message)
        }
      }

      const siteData = {
        ...formData,
        parser_config: parserConfig,
        scrape_frequency_hours: parseInt(formData.scrape_frequency_hours)
      }

      await onSave(siteData)
    } catch (err) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>{site.id ? 'Edit Site Config' : 'Add New Site'}</h3>
          <button className="close-button" onClick={onClose}>&times;</button>
        </div>

        {error && <div className="error">{error}</div>}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="site_name">Site Name *</label>
            <input
              type="text"
              id="site_name"
              name="site_name"
              value={formData.site_name}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="display_name">Display Name *</label>
            <input
              type="text"
              id="display_name"
              name="display_name"
              value={formData.display_name}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="base_url">Base URL *</label>
            <input
              type="url"
              id="base_url"
              name="base_url"
              value={formData.base_url}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="home_team">Home Team</label>
            <input
              type="text"
              id="home_team"
              name="home_team"
              value={formData.home_team}
              onChange={handleChange}
            />
          </div>

          <div className="form-group">
            <label htmlFor="parser_type">Parser Type *</label>
            <select
              id="parser_type"
              name="parser_type"
              value={formData.parser_type}
              onChange={handleChange}
              required
            >
              <option value="day_details">Day Details</option>
              <option value="day_details_parser1">Day Details Parser1</option>
              <option value="day_details_parser2">Day Details Parser2</option>
              <option value="month_based">Month Based</option>
              <option value="group_based">Group Based</option>
              <option value="external">External</option>
              <option value="custom">Custom</option>
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="parser_config">Parser Config (JSON)</label>
            <textarea
              id="parser_config"
              name="parser_config"
              value={parserConfigText}
              onChange={handleParserConfigChange}
              placeholder='{"url_template": "Calendar/?Month=%d&Year=%d"}'
            />
          </div>

          <div className="form-group">
            <label htmlFor="scrape_frequency_hours">Scrape Frequency (hours)</label>
            <input
              type="number"
              id="scrape_frequency_hours"
              name="scrape_frequency_hours"
              value={formData.scrape_frequency_hours}
              onChange={handleChange}
              min="1"
            />
          </div>

          <div className="form-group">
            <label htmlFor="notes">Notes</label>
            <textarea
              id="notes"
              name="notes"
              value={formData.notes}
              onChange={handleChange}
              style={{ minHeight: '60px' }}
            />
          </div>

          <div className="form-group-inline">
            <input
              type="checkbox"
              id="enabled"
              name="enabled"
              checked={formData.enabled}
              onChange={handleChange}
            />
            <label htmlFor="enabled">Enabled</label>
          </div>

          <div className="form-actions">
            <button type="button" className="btn btn-secondary" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className="btn btn-success" disabled={saving}>
              {saving ? 'Saving...' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default SiteEditModal
