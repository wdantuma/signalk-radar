import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent {
  title = 'radar-client';
}

if (typeof Worker !== 'undefined') {
  // Create a new
  const worker = new Worker(new URL('../workers/radar.worker', import.meta.url));
  worker.onmessage = ({ data }) => {
    //console.log(`page got message: ${data}`);
  };
} else {
  // Web Workers are not supported in this environment.
  // You should add a fallback so that your program still executes correctly.
}
