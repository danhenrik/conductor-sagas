const axios = require("axios");
const { faker } = require("@faker-js/faker");
const fs = require("fs");

const flightUrl = "http://localhost:3000/flights";
const hotelUrl = "http://localhost:3001/hotels";

const numFlights = 1;
const numHotels = 4;

async function createFlights() {
  const airlines = ["Latam", "Gol", "Azul", "American Airlines", "Delta"];
  const origins = [
    "São Paulo",
    "Rio de Janeiro",
    "Belo Horizonte",
    "Brasília",
    "Salvador",
  ];
  const destinations = ["New York", "London", "Paris", "Tokyo", "Dubai"];

  let ids = [];
  for (let i = 0; i < numFlights; i++) {
    const flightData = {
      airline: faker.helpers.arrayElement(airlines),
      origin: faker.helpers.arrayElement(origins), // Cidade de origem aleatória
      destination: faker.helpers.arrayElement(destinations), // Cidade de destino aleatória
      departureTime: "2023-10-01T10:00:00Z",
      arrivalTime: "2023-10-01T12:00:00Z",
      capacity: 320,
    };

    try {
      const response = await axios.post(flightUrl, flightData);
      var flightId = response.data.flightId;
      console.log(`"${flightId}" - Flight criado com sucesso!`);
      ids.push(flightId);
    } catch (error) {
      console.error(`Erro ao criar o voo ${flightId}:`, error);
    }
  }
  return ids;
}

async function createHotels() {
  let ids = [];
  for (let i = 0; i < numHotels; i++) {
    const hotelData = {
      name: faker.company.name(),
      location: `${faker.location.city()}, ${faker.location.state()}`,
      rating: faker.number.int({ min: 1, max: 5 }),
      roomsAvailable: 100,
    };

    try {
      const response = await axios.post(hotelUrl, hotelData);
      var hotelId = response.data.hotelId;
      console.log(`"${hotelId}" - Hotel criado com sucesso!`);
      ids.push(hotelId);
    } catch (error) {
      console.error(`Erro ao criar o hotel ${hotelId}:`, error);
    }
  }
  return ids;
}

function generateBatches(batchSize, flightIds, hotelIds) {
  const batches = [];
  let flightIndex = 0;
  let hotelIndex = 0;
  let seatNumber = 1;
  let roomNumber = 1;

  const batch = [];
  for (let j = 0; j < batchSize; j++) {
    const flightId = flightIds[flightIndex % flightIds.length];
    const hotelId = hotelIds[hotelIndex % hotelIds.length];

    batch.push({
      flightId: flightId,
      hotelId: hotelId,
      customerName: faker.person.firstName(),
      customerEmail: faker.internet.email(),
      seatNumber: seatNumber,
      checkInDate: faker.date.future().toISOString(),
      checkOutDate: faker.date.future().toISOString(),
      roomNumber: roomNumber,
    });

    seatNumber++;
    if (seatNumber > 400) {
      seatNumber = 1;
      flightIndex++;
    }

    roomNumber++;
    if (roomNumber > 100) {
      roomNumber = 1;
      hotelIndex++;
    }
  }
  batches.push(batch);
  return batches;
}

function saveBatch(batch) {
  const header =
    "flightId,hotelId,customerName,customerEmail,seatNumber,checkInDate,checkOutDate,roomNumber\n";
  const csv =
    header +
    batch
      .map((booking) => {
        return `${booking.flightId},${booking.hotelId},${booking.customerName},${booking.customerEmail},${booking.seatNumber},${booking.checkInDate},${booking.checkOutDate},${booking.roomNumber}`;
      })
      .join("\n");

  let filename = `batches/batch-${batchIndex}.csv`;
  fs.writeFileSync(filename, csv);
  console.log(`"${filename}" - Batch salvo!`);
  batchIndex++;
}

var batchIndex = 0;

(async () => {
  var requestCount = process.argv[2];
  if (!requestCount) {
    console.log(
      "node generate-test-mass.js <TotalRequestCount> <TotalBatchCount>"
    );
    process.exit(1);
  }

  var batchCount = process.argv[3];
  if (!batchCount) {
    console.log(
      "node generate-test-mass.js <TotalRequestCount> <TotalBatchCount>"
    );
    process.exit(1);
  }

  var batchSize = requestCount / batchCount;
  if (batchSize > 400) {
    console.log("Batch size must be less than 400");
    process.exit(1);
  }

  // Clean up batches directory
  fs.rmdirSync("batches", { recursive: true });
  fs.mkdirSync("batches");

  // Generate successful batches
  let successRequestCount = requestCount * (4 / 5);
  for (let i = 0; i < successRequestCount; i += batchSize) {
    let flightIds = await createFlights();
    let hotelIds = await createHotels();

    var batches = generateBatches(batchSize, flightIds, hotelIds);
    batches.forEach(saveBatch);
  }

  // Generate failing batches
  let failRequestCount = requestCount * (1 / 5);
  for (let i = 0; i < failRequestCount; i += batchSize) {
    let flightIds = await createFlights();
    let hotelIds = new Array(flightIds.length).fill("invalid-hotel-id");
    var batches = generateBatches(batchSize, flightIds, hotelIds);
    batches.forEach(saveBatch);
  }
})();