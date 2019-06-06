import React from 'react';
import Table from 'react-bootstrap/Table'
import Alert from 'react-bootstrap/Alert'

class WeatherTable extends React.Component {
  render() {
    let { city, weatherData, error } = this.props
    if (city.length > 0 && weatherData && weatherData.length > 0) {
      let weatherRows = weatherData.map((data) => {
        let details = data.error && data.error.length > 0 ? data.error : data.detailed_description
        return <tr key={data.source}><td>{data.source}</td><td>{data.temperature}&deg;C</td><td>{data.temperature_min}&deg;C</td><td>{data.temperature_max}&deg;C</td><td>{data.main_description}</td><td>{details}</td></tr>
      })
      return (
        <div>
          <h3>Weather Results for {this.props.city}</h3>
          <Table striped bordered hover variant="dark" responsive>
            <thead>
              <tr>
                <th>Source</th>
                <th>Temperature</th>
                <th>Temperature Min</th>
                <th>Temperature Max</th>
                <th>Conditions</th>
                <th>Details</th>
              </tr>
            </thead>
            <tbody>
              {weatherRows}
            </tbody>
          </Table>
        </div>
      )
    } else if (error && error.length > 0) {
      return <Alert key="error" variant="danger">Failed to get weather data: {error}</Alert>
    }
    return null
  }

}

export default WeatherTable;