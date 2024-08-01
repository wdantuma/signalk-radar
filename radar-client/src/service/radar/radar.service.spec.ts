import { TestBed } from '@angular/core/testing';

import { RadarService } from './radar.service';

describe('RadarService', () => {
  let service: RadarService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(RadarService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
