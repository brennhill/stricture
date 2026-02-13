// youtube-client-b08.ts — Handles "none" and "live" but forgets "upcoming".

type LiveBroadcastContentB08 = "none" | "live";
// BUG: Missing "upcoming" — a valid enum value per API spec

interface VideoSnippetB08 {
  title: string;
  liveBroadcastContent: LiveBroadcastContentB08;
  // ... other fields
}

class YouTubeClientB08 {
  private apiKey: string;
  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  formatVideoStatus(snippet: VideoSnippetB08): string {
    // BUG: exhaustive switch is missing "upcoming"
    switch (snippet.liveBroadcastContent) {
      case "none":
        return "On-demand video";
      case "live":
        return "Currently streaming";
      // Missing case "upcoming":
      //   return "Scheduled livestream";
      default:
        // At runtime, "upcoming" falls here and is treated as an unknown value
        return "Unknown status";
    }
  }

  isLive(snippet: VideoSnippetB08): boolean {
    // BUG: Only checks for "live", treats "upcoming" same as "none"
    return snippet.liveBroadcastContent === "live";
  }

  shouldShowLiveIndicator(snippet: VideoSnippetB08): boolean {
    // BUG: An upcoming stream SHOULD show a "Starting soon" indicator
    // but this only shows indicator for "live"
    return snippet.liveBroadcastContent === "live";
  }
}

async function testB08() {
  const client = new YouTubeClientB08("test-key");

  // Tests only cover "none" and "live"
  assert.equal(
    client.formatVideoStatus({ title: "Video", liveBroadcastContent: "none" }),
    "On-demand video"
  );
  assert.equal(
    client.formatVideoStatus({ title: "Stream", liveBroadcastContent: "live" }),
    "Currently streaming"
  );
  // BUG: No test for "upcoming"
  // If tested, it would return "Unknown status" instead of "Scheduled livestream"
}
