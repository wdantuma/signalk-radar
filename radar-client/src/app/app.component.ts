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
import {Icon,Style} from 'ol/style'
import { RadarService } from '../service/radar/radar.service'
import { fromLonLat } from 'ol/proj'


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

  constructor(private radarService:RadarService) { this.map = null; }
  ngOnInit(): void {

   const shipLocation = fromLonLat(this.radarService.GetShipLocation(), "EPSG:3857")

    const radarLayer = new ImageLayer({
      source:this.radarService.createRadarSource(0)
    })

    const shipFeature =new Feature(new Point(shipLocation))
    const shipStyle =new Style({
      image: new Icon({
       src:'img/ship_red.png',
       rotation:2.04,
       scale:0.5
      })
    })
    shipFeature.setStyle(shipStyle)
    const shipSource = new VectorSource<FeatureLike>({features:[shipFeature]})
    const shipLayer = new VectorLayer({ source: shipSource })

   
    this.map = new OlMap({
      view: new View({
        center: shipLocation,
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
  }
}

