/// <reference lib="webworker" />

import { Radar } from './radar.model'
import { RadarMessage } from './RadarMessage'
import { Color } from 'ol/color'

enum BlobColour {
  BLOB_NONE = 0,
  BLOB_HISTORY_0,
  BLOB_HISTORY_1,
  BLOB_HISTORY_2,
  BLOB_HISTORY_3,
  BLOB_HISTORY_4,
  BLOB_HISTORY_5,
  BLOB_HISTORY_6,
  BLOB_HISTORY_7,
  BLOB_HISTORY_8,
  BLOB_HISTORY_9,
  BLOB_HISTORY_10,
  BLOB_HISTORY_11,
  BLOB_HISTORY_12,
  BLOB_HISTORY_13,
  BLOB_HISTORY_14,
  BLOB_HISTORY_15,
  BLOB_HISTORY_16,
  BLOB_HISTORY_17,
  BLOB_HISTORY_18,
  BLOB_HISTORY_19,
  BLOB_HISTORY_20,
  BLOB_HISTORY_21,
  BLOB_HISTORY_22,
  BLOB_HISTORY_23,
  BLOB_HISTORY_24,
  BLOB_HISTORY_25,
  BLOB_HISTORY_26,
  BLOB_HISTORY_27,
  BLOB_HISTORY_28,
  BLOB_HISTORY_29,
  BLOB_HISTORY_30,
  BLOB_HISTORY_31,
  BLOB_WEAK,
  BLOB_INTERMEDIATE,
  BLOB_STRONG,
  BLOB_DOPPLER_RECEDING,
  BLOB_DOPPLER_APPROACHING
};

const BLOB_COLOURS = (BlobColour.BLOB_DOPPLER_APPROACHING + 1)
const BLOB_HISTORY_MAX = (BlobColour.BLOB_HISTORY_31)
let m_colour_map = new Map<number, BlobColour>()
let m_colour_map_rgb = new Map<BlobColour, Color>()
const thresholdRed = 255
const thresholdGreen = 255
const thresholdBlue = 255
let Heading = 0

function computeColourMap(doppler_states: number) {
  for (let i = 0; i <= 255; i++) {
    if (i == 255 && doppler_states > 0) {
      m_colour_map.set(i, BlobColour.BLOB_DOPPLER_APPROACHING)
    } else if ((i == 255 - 1) && doppler_states == 1) {
      m_colour_map.set(i, BlobColour.BLOB_DOPPLER_RECEDING)
    } else if (i >= thresholdRed) {
      m_colour_map.set(i, BlobColour.BLOB_STRONG)
    } else if (i >= thresholdGreen) {
      m_colour_map.set(i, BlobColour.BLOB_INTERMEDIATE)
    } else if (i >= thresholdBlue && i > BLOB_HISTORY_MAX) {
      m_colour_map.set(i, BlobColour.BLOB_WEAK)
    } else {
      m_colour_map.set(i, BlobColour.BLOB_NONE)
    }
  }
  for (let i = 0; i < BLOB_COLOURS; i++) {
    m_colour_map_rgb.set(i, [0, 0, 0, 0])
  }
  m_colour_map_rgb.set(BlobColour.BLOB_DOPPLER_APPROACHING, [255, 255, 0, 1]) // yellow
  m_colour_map_rgb.set(BlobColour.BLOB_DOPPLER_RECEDING, [0, 255, 255, 1]) // cyan
  m_colour_map_rgb.set(BlobColour.BLOB_STRONG, [255, 0, 0, 1]) // red
  m_colour_map_rgb.set(BlobColour.BLOB_INTERMEDIATE, [0, 255, 0, 1]) // green
  m_colour_map_rgb.set(BlobColour.BLOB_WEAK, [0, 0, 255, 1]) // blue
}


addEventListener('message', (event) => {
  if (event.data.heading) {
    Heading = event.data.heading;
  }
  if (event.data.canvas) {
    computeColourMap(0)
    const radarCanvas = event.data.canvas
    const radar = event.data.radar as Radar
    const ctxWorker = radarCanvas.getContext("2d") as CanvasRenderingContext2D;
    const pixel = ctxWorker.createImageData(1, 1)
    const pixelData = pixel.data
    pixelData[0] = 0
    pixelData[1] = 0
    pixelData[2] = 0
    pixelData[3] = 255

    let x: number[] = []
    let y: number[] = []

    const cx = radarCanvas.width / 2
    const cy = radarCanvas.height / 2

    for (let a = 0; a < radar.spokes; a++) {
      for (let r = 0; r < radar.maxSpokeLen; r++) {
        const angle = a * ((2 * Math.PI) / radar.spokes)
        const radius = r * ((radarCanvas.width / 2) / radar.maxSpokeLen)
        const x1 = Math.round(cx + radius * Math.cos(angle))
        const y1 = Math.round(cy + radius * Math.sin(angle))
        x[a * radar.maxSpokeLen + r] = x1
        y[a * radar.maxSpokeLen + r] = y1
      }
    }


    function connect() {
      const socket = new WebSocket(radar.streamUrl);
      socket.binaryType = "arraybuffer"
  
      let lastRange = 0
  
      socket.onmessage = (event) => {
        let message = RadarMessage.deserialize(event.data)
        if (lastRange != message.spoke.range) {
          lastRange = message.spoke.range
          postMessage({ range: message.spoke.range })
        }
  
        let angle = message.spoke.angle
        let h = Heading-90
        if (h<0) {
          h+=360
        }
        angle += Math.round((h) / (360 / radar.spokes)) // add heading
        angle = angle % radar.spokes
  
        // 2D context based draw implementation maybe to webgl context
        if (Date.now() - message.spoke.time < 1000) { // drop old spokes
          // clear spoke in front
          let clearangle1 = angle + 1 % radar.spokes
          let clearangle2 = angle + 4 % radar.spokes
          ctxWorker.moveTo
          ctxWorker.save()
          ctxWorker.beginPath()
          ctxWorker.strokeStyle = "#00000000"
          ctxWorker.moveTo(x[clearangle1 * radar.maxSpokeLen + 0], y[clearangle1 * radar.maxSpokeLen + 0])
          ctxWorker.lineTo(x[clearangle1 * radar.maxSpokeLen + radar.maxSpokeLen - 1], y[clearangle1 * radar.maxSpokeLen + radar.maxSpokeLen - 1])
          ctxWorker.lineTo(x[clearangle2 * radar.maxSpokeLen + radar.maxSpokeLen - 1], y[clearangle2 * radar.maxSpokeLen + radar.maxSpokeLen - 1])
          ctxWorker.closePath()
          ctxWorker.stroke()
          ctxWorker.clip()
          ctxWorker.clearRect(0, 0, radarCanvas.width, radarCanvas.height);
          ctxWorker.restore()
  
          // draw current spoke
          for (let i = 0; i < message.spoke.data.length; i++) {
            let ci = m_colour_map.get(message.spoke.data[i])
            if (ci != BlobColour.BLOB_NONE) {
              let color = m_colour_map_rgb.get(ci as BlobColour)
              if (color) {
                let x1 = x[angle * radar.maxSpokeLen + i]
                let y1 = y[angle * radar.maxSpokeLen + i]
                pixelData[0] = color[0]
                pixelData[1] = color[1]
                pixelData[2] = color[2]
                pixelData[3] = color[3] * 255
                ctxWorker.putImageData(pixel, x1, y1)
              }
            }
          }
          postMessage({ redraw: true })
  
        }
      }

      socket.onopen = (event) => {
        console.log(`Radar ${radar.label} connected`)
      }
  
      socket.onclose = (event) => {
        console.log(`Radar ${radar.label} disconnected retry in 3 seconds`);
        ctxWorker.clearRect(0, 0, radarCanvas.width, radarCanvas.height);
        postMessage({ redraw: true })
        setTimeout(function() {
          connect();
        }, 3000);
      }      
  
      socket.onerror = (event) => {
        ctxWorker.clearRect(0, 0, radarCanvas.width, radarCanvas.height);
        postMessage({ redraw: true })
        console.error(`Error on radar ${radar.label} stopping`)
      }
    }
    connect();
  }
});
