import { Component, OnInit } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import OlMap from 'ol/Map';
import Coordinate from 'ol/coordinate'
import View from 'ol/View';
import TileLayer from 'ol/layer/Tile';
import ImageLayer from 'ol/layer/Image'
import VectorLayer from 'ol/layer/Vector'
import ImageCanvas from 'ol/source/ImageCanvas'
import VectorSource from 'ol/source/Vector'
import OSM from 'ol/source/OSM';
import { fromLonLat } from 'ol/proj'
import Projection from 'ol/proj/Projection'
import { transform } from 'ol/proj'
import { boundingExtent } from 'ol/extent'
import Point from 'ol/geom/Point'
import Feature from 'ol/Feature'
import Circle from 'ol/geom/Circle'
import {Icon,Style} from 'ol/style'

const worker = new Worker(new URL('../workers/radar.worker', import.meta.url));

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

  constructor() { this.map = null; }
  ngOnInit(): void {

    // location boat  -60.841278 11.157449, -60.841278
    const center = [-60.841983, 11.157833]
    const projectedCenter = fromLonLat(center, "EPSG:3857")
    const extent = boundingExtent([[projectedCenter[0] - 100, projectedCenter[1] - 100], [projectedCenter[0] + 100, projectedCenter[1] + 100]])

    const projection = new Projection({
      code: 'radar',
      units: 'm',
    });

    const radarCanvas = document.createElement("canvas")
    radarCanvas.setAttribute('width', '1382px')
    radarCanvas.setAttribute('height', '1382px')

    const offscreenRdarcanvas = radarCanvas.transferControlToOffscreen()

    const radarLayer = new ImageLayer({
      extent: extent,
      maxZoom: 17,
      minZoom: 10,
      source: new ImageCanvas({
        projection: projection,
        canvasFunction: (extent, resolution, ratio, size, projection) => {
          return radarCanvas
        }
      })
    })

    const shipFeature =new Feature(new Point(projectedCenter))
    const shipStyle =new Style({
      image: new Icon({
       src:'img/ship_red.png',
       rotation:2.04,
       scale:0.5
      })
    })
    shipFeature.setStyle(shipStyle)
    const shipSource = new VectorSource({features:[shipFeature]})
    const shipLayer = new VectorLayer({ source: shipSource })

    worker.postMessage({ canvas: offscreenRdarcanvas }, [offscreenRdarcanvas]);
    worker.onmessage = (event) => {
      if (event.data.redraw) {
        radarLayer.getSource()?.changed()
      } else if (event.data.range) {
        let rangeCircel = new Circle(transform(center, 'EPSG:4326', 'EPSG:3857'), 3000)
        radarLayer.setExtent(rangeCircel.getExtent())
      }
    }

    this.map = new OlMap({
      view: new View({
        center: projectedCenter,
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

