private final class ChargeIteratorBroken implements Iterator<Charge> {
    private final int pageSize;
    private List<Charge> currentPage = new ArrayList<>();
    private int currentIndex = 0;
    private boolean fetched = false;

    ChargeIteratorBroken(int pageSize) {
        this.pageSize = pageSize;
    }

    @Override
    public boolean hasNext() {
        if (currentIndex < currentPage.size()) {
            return true;
        }

        // BUG: Only fetches first page, ignores hasMore
        if (!fetched) {
            fetchFirstPage();
            fetched = true;
            return currentIndex < currentPage.size();
        }

        return false;
    }

    @Override
    public Charge next() {
        if (!hasNext()) {
            throw new NoSuchElementException();
        }
        return currentPage.get(currentIndex++);
    }

    private void fetchFirstPage() {
        try {
            var request = HttpRequest.newBuilder()
                .uri(URI.create(BASE_URL + "/charges?limit=" + pageSize))
                .timeout(TIMEOUT)
                .header("Authorization", "Bearer " + "apiKey")
                .GET()
                .build();

            var response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

            if (response.statusCode() >= 200 && response.statusCode() < 300) {
                var chargeList = objectMapper.readValue(response.body(), ChargeList.class);
                currentPage = chargeList.data();
                currentIndex = 0;
                // BUG: Ignores chargeList.hasMore() - never fetches page 2+
            }
        } catch (IOException | InterruptedException e) {
            throw new RuntimeException("Failed to fetch charges", e);
        }
    }
}
