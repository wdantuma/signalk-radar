import { Component, OnInit } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import OlMap from 'ol/Map';
import View from 'ol/View';
import TileLayer from 'ol/layer/Tile';
import ImageLayer from 'ol/layer/Image'
import VectorLayer from 'ol/layer/Vector'
import VectorSource from 'ol/source/Vector'
import OSM from 'ol/source/OSM';
import Point from 'ol/geom/Point'
import Feature, { FeatureLike } from 'ol/Feature'
import { Icon, Style } from 'ol/style'
import { RadarService } from '../service/radar/radar.service'
import { fromLonLat } from 'ol/proj'
import { Radar } from '../service/radar/radar.model';
import { BehaviorSubject} from 'rxjs'
import { ShipState } from '../service/radar/ship-state.model ';
import { DragRotateAndZoom,defaults} from 'ol/interaction'


@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent implements OnInit {
  title = 'radar-client';

  map: OlMap | null;

  constructor(private radarService: RadarService) { this.map = null; }
  ngOnInit(): void {



    let start = [-60.885435, 11.175774]
    let end = [-60.841983, 11.157833]


    let subject = new BehaviorSubject<ShipState>({location:start,heading:117});

    let step=0;
    let res = 1000
    setInterval(() => {
      let xd = (end[0]-start[0])/res
      let yd = (end[1]-start[1])/res    
      step++
      if(step>res-1) {
        step=0
      }

      subject.next( {location: [start[0]+step*xd,start[1]+step*yd],heading:117});
    }, 100);


    const radarLayer = new ImageLayer({
    })


    const shipSource = new VectorSource<FeatureLike>()
    const shipLayer = new VectorLayer({ source: shipSource })

    subject.subscribe((location) => {
      let shipLocation = fromLonLat(location.location)
      let angle = location.heading*(Math.PI/180)
      const shipFeature = new Feature(new Point(shipLocation))
      const shipStyle = new Style({
        image: new Icon({
          src: 'img/ship_red.png',
          rotation: angle,
          rotateWithView: true,
          scale: 0.5
        })
      })  
      shipFeature.setStyle(shipStyle)
      shipSource.clear();
      shipSource.addFeature(shipFeature)      
    })


    this.map = new OlMap({
      interactions: defaults().extend([new DragRotateAndZoom()]),
      view: new View({
        center: fromLonLat(start),
        zoom: 16,
      }),
      layers: [
        new TileLayer({
          source: new OSM(),
        }),
        radarLayer,
        shipLayer
      ],
      target: 'ol-map'
    });

    this.radarService.Connect("http://localhost:3001").then(() => {
      let radars = this.radarService.GetRadars() 
      let radar = radars.get(radars.keys().next().value);
      if(radar) {
        radarLayer.setSource(this.radarService.CreateRadarSource(radar,subject))
      }
    })
  }
}

