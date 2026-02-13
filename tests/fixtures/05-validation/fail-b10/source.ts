// youtube-client-b10.ts — Duration "PT1H2M3S" treated as plain string, no ISO 8601 parsing.

interface VideoDurationInfoB10 {
  videoId: string;
  duration: string;  // Stored as raw string, never validated or parsed
}

class YouTubeClientB10 {
  private apiKey: string;
  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getVideoDurations(videoIds: string[]): Promise<VideoDurationInfoB10[]> {
    const response = await fetch(
      `https://www.googleapis.com/youtube/v3/videos?part=contentDetails&id=${videoIds.join(",")}&key=${this.apiKey}`
    );
    if (!response.ok) throw new Error(`API error: ${response.status}`);
    const data = (await response.json()) as VideoListResponse;

    return data.items.map((video) => ({
      videoId: video.id,
      // BUG: Duration stored as raw string with no format validation.
      // If the API ever returns a malformed duration, it is silently accepted.
      duration: video.contentDetails!.duration,
    }));
  }

  // BUG: Tries to display duration as "H:MM:SS" but does string manipulation
  // instead of proper ISO 8601 parsing
  formatDuration(duration: string): string {
    // This naive approach fails for many valid ISO 8601 durations:
    //   "PT3M33S"  -> replaces correctly by accident
    //   "PT1H2M3S" -> produces "1:2:3" instead of "1:02:03"
    //   "PT45S"    -> produces "45" instead of "0:45"
    //   "PT1H"     -> produces "1" instead of "1:00:00"
    return duration
      .replace("PT", "")
      .replace("H", ":")
      .replace("M", ":")
      .replace("S", "");
  }

  // BUG: No validation that duration matches ISO 8601 pattern
  isShortVideo(duration: string): boolean {
    // Tries to check if video is under 60 seconds, but uses indexOf
    // "PT59S" -> no "M" or "H", returns true (correct by accident)
    // "PT1M0S" -> has "M", returns false (correct by accident)
    // "PT0M59S" -> has "M", returns false (WRONG — 59 seconds IS short)
    return !duration.includes("H") && !duration.includes("M");
  }
}

async function testB10() {
  const client = new YouTubeClientB10("test-key");

  // Only tests the happy-path format
  assert.equal(client.formatDuration("PT3M33S"), "3:33");

  // BUG: Does not test:
  //   formatDuration("PT1H2M3S")  -> "1:2:3" (wrong, should be "1:02:03")
  //   formatDuration("PT45S")     -> "45" (wrong, should be "0:45")
  //   formatDuration("PT1H")      -> "1" (wrong, should be "1:00:00")
  //   isShortVideo("PT0M59S")     -> false (wrong, 59 seconds is short)
  //   formatDuration("invalid")   -> "invalid" (should throw)
}
