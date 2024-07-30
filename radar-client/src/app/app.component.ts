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
import  Projection  from 'ol/proj/Projection'
import { boundingExtent } from 'ol/extent'
import Point from 'ol/geom/Point'
import Feature from 'ol/Feature'

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
    const center = [-60.841278, 11.157449]
    const projectedCenter = fromLonLat(center, "EPSG:3857")
    const extent = boundingExtent([[projectedCenter[0] - 100, projectedCenter[1] - 100], [projectedCenter[0] + 100, projectedCenter[1] + 100]])

    const projection = new Projection({
      code: 'radar',
      units: 'pixels',
      extent: extent,
    });

    const radarCanvas = document.createElement("canvas")
    radarCanvas.setAttribute('width','200')
    radarCanvas.setAttribute('height','200')

    const offscreenRdarcanvas = radarCanvas.transferControlToOffscreen()

    const radarLayer = new ImageLayer({
      extent: extent,
      source: new ImageCanvas({
        projection:projection,
        canvasFunction: (extent, resolution, ratio, size, projection) => {
          return radarCanvas
        }
      })
    })

    const shipSource = new VectorSource()
    const shipLayer = new VectorLayer({ source: shipSource })
    shipSource.addFeature(new Feature(new Point(projectedCenter)))


    worker.postMessage({ canvas: offscreenRdarcanvas }, [offscreenRdarcanvas]);
    worker.onmessage = (event) => {
      if (event.data.redraw) {
        radarLayer.getSource()?.changed()
      } else if (event.data.range) {
        const extent = boundingExtent([[projectedCenter[0] - event.data.range, projectedCenter[1] - event.data.range], [projectedCenter[0] + event.data.range, projectedCenter[1] + event.data.range]])
        radarLayer.setExtent(extent)
      }
    }

    this.map = new OlMap({
      view: new View({
        center: projectedCenter,
        zoom: 16,
        maxZoom:20
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

