import { Injectable } from '@angular/core';
import ImageSource from 'ol/source/Image'
import { fromLonLat } from 'ol/proj'
import Projection from 'ol/proj/Projection'
import Circle from 'ol/geom/Circle'
import {createLoader} from 'ol/source/static'
import { Coordinate } from 'ol/coordinate';

const worker = new Worker(new URL('./radar.worker', import.meta.url));

@Injectable({
  providedIn: 'root'
})
export class RadarService {

  constructor() { }

  public GetShipLocation():Coordinate {
    const shipLocation = [-60.841983, 11.157833]
    return shipLocation
  }

  public createRadarSource(radar :number):ImageSource {

     // location boat  -60.841278 11.157449, -60.841278
     
     const projectedCenter = fromLonLat(this.GetShipLocation(), "EPSG:3857")
 
     const projection = new Projection({
       code: 'radar',
       units: 'm',      
     });
 
     let range = 3710
     let rangeExtent =  new Circle(projectedCenter, range).getExtent()
 
     const radarCanvas = document.createElement("canvas")
     radarCanvas.width = 1410 // twice 
     radarCanvas.height =1410
 
     const offscreenRdarcanvas = radarCanvas.transferControlToOffscreen()

     let radarSource = new ImageSource({
      projection:projection,
      loader: createLoader({imageExtent:rangeExtent,url:"",load:() => {
         return Promise.resolve(radarCanvas)
      }})
     })  

    worker.postMessage({ canvas: offscreenRdarcanvas,radar:radar }, [offscreenRdarcanvas]);
    worker.onmessage = (event) => {
      if (event.data.redraw) {
        radarSource.refresh()
      } else if (event.data.range) {   
        rangeExtent = new Circle(projectedCenter, event.data.range).getExtent()
      }
    }
    return radarSource
  }
}
