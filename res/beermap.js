var map;

function initMap() {
  fetch("pubs.json")
    .then(function(response) {
      return response.json();
    })
    .then(function(mapData) {
      console.log(mapData.bounds);
      startMap(mapData);
    })
}

function startMap(mapData) {

  var styles = [
      {
          "featureType": "administrative",
          "elementType": "labels.text.fill",
          "stylers": [
              {
                  "color": "#444444"
              }
          ]
      },
      {
          "featureType": "landscape",
          "elementType": "all",
          "stylers": [
              {
                  "color": "#f2f2f2"
              }
          ]
      },
      {
          "featureType": "poi",
          "elementType": "all",
          "stylers": [
              {
                  "visibility": "on"
              }
          ]
      },
      {
          "featureType": "road",
          "elementType": "all",
          "stylers": [
              {
                  "saturation": -100
              },
              {
                  "lightness": 45
              }
          ]
      },
      {
          "featureType": "road.highway",
          "elementType": "all",
          "stylers": [
              {
                  "visibility": "simplified"
              }
          ]
      },
      {
          "featureType": "road.arterial",
          "elementType": "labels.icon",
          "stylers": [
              {
                  "visibility": "off"
              }
          ]
      },
      {
          "featureType": "transit",
          "elementType": "all",
          "stylers": [
              {
                  "visibility": "on"
              }
          ]
      },
      {
          "featureType": "water",
          "elementType": "all",
          "stylers": [
              {
                  "color": "#46bcec"
              },
              {
                  "visibility": "on"
              }
          ]
      }
  ];

  var mapElement = document.getElementById('map');
  map = new google.maps.Map(mapElement, {
    styles: styles,
  });
  map.fitBounds(mapData.bounds);

  var transitLayer = new google.maps.TransitLayer();
  transitLayer.setMap(map);

  var infowindow = new google.maps.InfoWindow();

  var markers = mapData.pubs.map(function(p) {
    var color;
    if (p.closed) {
      color = "#f00";
    } else if (p.visited) {
      color = "#0c0";
    } else {
      color = "#44f";
    }
    var marker = new google.maps.Marker({
        position: new google.maps.LatLng(p.lat, p.lng),
        icon: svgCircleIcon(color, '#fff', p.label)
    });
    marker.addListener('click', function() {
      infowindow.setContent(p.content);
      infowindow.open(map, marker);
    });
    return marker;
  });

  markers.forEach(function(marker) {
    marker.setMap(map);
  })
}

function svgCircleIcon(circleFill, labelColor, label) {
  var template = [
      '<?xml version="1.0"?>',
          '<svg width="26px" height="26px" viewBox="0 0 100 100" version="1.1" xmlns="http://www.w3.org/2000/svg">',
              '<circle stroke="{{fill}}" stroke-width="10" stroke-opacity="10%" fill="{{fill}}" cx="50" cy="50" r="35"/>',
              '<text x="50" y="50" text-anchor="middle" alignment-baseline="central" font-family="Roboto" font-size="40" fill="{{labelFill}}">{{label}}</text>',
          '</svg>'
      ].join('\n');
  var svg = replaceAll(template, "{{fill}}", circleFill);
  svg = replaceAll(svg, "{{labelFill}}", labelColor);
  svg = replaceAll(svg, "{{label}}", label);
  return {
    url: 'data:image/svg+xml;charset=UTF-8,' + encodeURIComponent(svg),
    scaledSize: new google.maps.Size(30, 30)
  };
}

function replaceAll(target, search, replacement) {
  return target.replace(new RegExp(search, 'g'), replacement);
}
