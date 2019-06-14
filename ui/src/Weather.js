import React from 'react';
import WeatherTable from './WeatherTable';
import InputGroup from 'react-bootstrap/InputGroup'
import Dropdown from 'react-bootstrap/Dropdown';
import Button from 'react-bootstrap/Button';
import Form from 'react-bootstrap/Form'

class Weather extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      city: "",
      sources: [],
      selectedSource: "Defaults",
      validated: false,
      currentWeather: {city: "", data: [], error: ""}
    }
    this.handleChange = this.handleChange.bind(this)
    this.handleSubmit = this.handleSubmit.bind(this)
    this.handleSelect = this.handleSelect.bind(this)
    this.getWeather = this.getWeather.bind(this)
  }

  componentDidMount() {
    fetch('http://localhost:8080/v1/backends').then(res => res.json())
      .then((data) => {
        let sources = data.backends
        sources.unshift("Defaults")
        this.setState({ sources: sources })
      })
      .catch(() => {
        let sources = []
        sources.unshift("Defaults")
        this.setState({ sources: sources })
      })
  }

  render() {
    let dropDownItems = this.state.sources.map((source) =>
      <Dropdown.Item key={source} eventKey={source}>{source}</Dropdown.Item>
    )
    return (
      <div>
        <h1>Weather</h1>
        <Form
          onSubmit={this.handleSubmit}
          noValidate
          validated={this.state.validated}
        >

          <InputGroup>
            <Form.Control
              size="lg"
              autoFocus
              as="input"
              placeholder="City..."
              aria-label="City"
              aria-describedby="basic-addon2"
              onChange={this.handleChange}
              value={this.state.city}
              type="text"
              required
            />
            <InputGroup.Append>

              <Dropdown
                value={this.state.selectedSource}
                onSelect={this.handleSelect}
              >
                <Dropdown.Toggle variant="outline-secondary" id="get-weather" size="lg">
                  {this.state.selectedSource}
                </Dropdown.Toggle>
                <Dropdown.Menu>
                  {dropDownItems}
                </Dropdown.Menu>
              </Dropdown>
              <Button
                type="submit"
                size="lg"
                variant="primary"
              >
                Get weather!
            </Button>
            </InputGroup.Append>

          </InputGroup>

        </Form>
        <WeatherTable city={this.state.currentWeather.city} weatherData={this.state.currentWeather.data} error={this.state.currentWeather.error} />
      </div>
    )
  }

  handleChange(e) {
    this.setState({ city: e.target.value })
  }

  handleSelect(selectedValue) {
    this.setState({ selectedSource: selectedValue })
  }

  handleSubmit(e) {
    e.preventDefault();
    const form = e.currentTarget;
    if (form.checkValidity() === false) {
      e.stopPropagation();
    }

    this.setState(state => ({
      validated: true,
      city: this.state.city,
      selectedSource: this.state.selectedSource
    }))
    this.getWeather()
  }

  getWeather() {
    let {city, selectedSource } = this.state
    let uri = 'http://localhost:8080/v1/weather/' + city
    if (selectedSource !== "Defaults") {
      uri = uri + "?backend=" + selectedSource
    }
    fetch(uri).then(res => res.json())
      .then((data) => {
        this.setState({
          currentWeather: data
        })
      })
      .catch(() => {
        this.setState({ currentWeather: {city: "",data: []}, error: "Timeout fetching data from backend" })
      })
  }

}

export default Weather;