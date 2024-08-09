import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import ImageSource from 'ol/source/Image'
import { fromLonLat } from 'ol/proj'
import Projection from 'ol/proj/Projection'
import Circle from 'ol/geom/Circle'
import { createLoader } from 'ol/source/static'
import { Coordinate } from 'ol/coordinate';
import { Radar } from './radar.model';
import { ShipState } from './ship-state.model '
import { firstValueFrom, Observable } from 'rxjs'
import {createEmpty} from 'ol/extent'

@Injectable({
  providedIn: 'root'
})
export class RadarService {

  constructor(private http: HttpClient) { }

  private radarServerUrl?: string;
  private radars: Map<string,Radar> = new Map<string,Radar>();


  public async Connect(radarServerUrl: string) {
    this.radarServerUrl = radarServerUrl;
    this.radars = await firstValueFrom(this.http.get<Map<string,Radar>>(`${this.radarServerUrl}/v1/api/radars`))
  }

  public GetRadars(): Map<string,Radar> {
    return this.radars;
  }

  public CreateRadarSource(radar: Radar, shipState: Observable<ShipState>): ImageSource {

    let range = 0
    let location: Coordinate = [0, 0]
    let rangeExtent = createEmpty();

    function UpdateExtent(location: Coordinate, range: number) {
      let center = fromLonLat(location, "EPSG:3857")
      let extent = new Circle(center, range).getExtent()
      rangeExtent[0] = extent[0]
      rangeExtent[1] = extent[1]
      rangeExtent[2] = extent[2]
      rangeExtent[3] = extent[3]
    }

    UpdateExtent(location, range)

    const projection = new Projection({
      code: 'radar',
      units: 'm',
    });

    //

    const radarCanvas = document.createElement("canvas")
    radarCanvas.width = 2 * radar.maxSpokeLen
    radarCanvas.height = 2 * radar.maxSpokeLen

    const offscreenRdarcanvas = radarCanvas.transferControlToOffscreen()

    let radarSource = new ImageSource({
      projection: projection,
      loader: createLoader({
        imageExtent: rangeExtent, url: "", load: () => {
          return Promise.resolve(radarCanvas)
        }
      })
    })

    const worker = new Worker(new URL('./radar.worker', import.meta.url));
    worker.postMessage({ canvas: offscreenRdarcanvas, radar: radar }, [offscreenRdarcanvas]);
    worker.onmessage = (event) => {
      if (event.data.redraw) {
        radarSource.refresh()
      } else if (event.data.range) {
        range = event.data.range;
        UpdateExtent(location, range);
        radarSource.refresh()
      }
    }
    shipState.subscribe((state) => {
      location = state.location;
      UpdateExtent(location, range);
      worker.postMessage({ heading: state.heading });
      radarSource.refresh()
    })
    return radarSource
  }
}
