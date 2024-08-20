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
const thresholdRed = 200
const thresholdGreen = 100
const thresholdBlue = 32

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
    const radarOnScreenCanvas = event.data.canvas
    const radarOnScreenContext = radarOnScreenCanvas.getContext("2d")
    const radarCanvas = new OffscreenCanvas(radarOnScreenCanvas.width, radarOnScreenCanvas.height)
    const radar = event.data.radar as Radar
    const radarContext = radarCanvas.getContext("2d");// as CanvasRenderingContext2D;

    const pixel = radarContext.createImageData(1, 1)
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

    function ToNorthAngle(angle:number) :number {
      let h = Heading - 90
      if (h < 0) {
        h += 360
      }
      angle += Math.round((h) / (360 / radar.spokes)) // add heading
      angle = angle % radar.spokes
      return angle
    }


    function connect() {
      const socket = new WebSocket(radar.streamUrl);
      socket.binaryType = "arraybuffer"

      let lastRange = 0

      socket.onmessage = (event) => {
        let message = RadarMessage.deserialize(event.data)
        if (message.spokes.length > 0) {
          let shift = Date.now() - message.spokes[0].time
          if (shift > 800) {
            // drop old packets
            return
          }
        }
        let clearangle1 = ToNorthAngle(message.spokes[0].angle % radar.spokes)
        let clearangle2 = ToNorthAngle(message.spokes[message.spokes.length-1].angle % radar.spokes)
        radarContext.save()
        radarContext.beginPath()
        radarContext.strokeStyle = "#00000000"
        radarContext.moveTo(x[0], y[0])
        radarContext.lineTo(x[clearangle1 * radar.maxSpokeLen + radar.maxSpokeLen - 1], y[clearangle1 * radar.maxSpokeLen + radar.maxSpokeLen - 1])
        radarContext.lineTo(x[clearangle2 * radar.maxSpokeLen + radar.maxSpokeLen - 1], y[clearangle2 * radar.maxSpokeLen + radar.maxSpokeLen - 1])
        radarContext.closePath()
        radarContext.stroke()
        radarContext.clip()
        radarContext.clearRect(0, 0, radarCanvas.width, radarCanvas.height);
        radarContext.restore()
        let spokeImageData = radarContext.getImageData(0, 0, radarCanvas.width, radarCanvas.height)
        for (let si = 0; si < message.spokes.length; si++) {
          let spoke = message.spokes[si]

          if (lastRange != spoke.range) {
            radarContext.clearRect(0, 0, radarCanvas.width, radarCanvas.height);
            spokeImageData = radarContext.getImageData(0, 0, radarCanvas.width, radarCanvas.height)
            lastRange = spoke.range
            postMessage({ range: spoke.range })
          }

          let angle = ToNorthAngle(spoke.angle)

          // 2D context based draw implementation maybe to webgl context

          // draw current spoke


          for (let i = 0; i < spoke.data.length; i++) {
            let x1 = x[angle * radar.maxSpokeLen + i]
            let y1 = y[angle * radar.maxSpokeLen + i]
            let index = (y1 * spokeImageData.width) + x1
            index = index * 4
            let ci = m_colour_map.get(spoke.data[i])
            if (ci != BlobColour.BLOB_NONE) {
              let color = m_colour_map_rgb.get(ci as BlobColour)
              if (color) {

                spokeImageData.data[index] = color[0]
                spokeImageData.data[index + 1] = color[1]
                spokeImageData.data[index + 2] = color[2]
                spokeImageData.data[index + 3] = color[3] * 255
              }
            } else {
              spokeImageData.data[index + 3] = 0
            }
          }

        }
        radarContext.putImageData(spokeImageData, 0, 0)
        radarOnScreenContext.clearRect(0, 0, radarCanvas.width, radarCanvas.height);
        radarOnScreenContext.drawImage(radarCanvas, 0, 0)
        postMessage({ redraw: true })
      }

      socket.onopen = (event) => {
        console.log(`Radar ${radar.name} connected`)
      }

      socket.onclose = (event) => {
        console.log(`Radar ${radar.name} disconnected retry in 3 seconds`);
        radarContext.clearRect(0, 0, radarCanvas.width, radarCanvas.height);
        radarOnScreenContext.clearRect(0, 0, radarCanvas.width, radarCanvas.height);

        setTimeout(function () {
          connect();
        }, 3000);
      }

      socket.onerror = (event) => {
        radarContext.clearRect(0, 0, radarCanvas.width, radarCanvas.height);
        radarOnScreenContext.clearRect(0, 0, radarCanvas.width, radarCanvas.height);
        postMessage({ redraw: true })
        console.error(`Error on radar ${radar.name} stopping`)
      }
    }
    connect();
  }
});
