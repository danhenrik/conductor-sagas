const axios = require("axios");
const { v4: uuidv4 } = require("uuid");
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
    const flightId = uuidv4();
    const flightData = {
      id: flightId,
      airline: faker.helpers.arrayElement(airlines),
      origin: faker.helpers.arrayElement(origins), // Cidade de origem aleatória
      destination: faker.helpers.arrayElement(destinations), // Cidade de destino aleatória
      departureTime: "2023-10-01T10:00:00Z",
      arrivalTime: "2023-10-01T12:00:00Z",
      capacity: 320,
    };

    try {
      const response = await axios.post(flightUrl, flightData);
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
    const hotelId = uuidv4();
    const hotelData = {
      id: hotelId,
      name: faker.company.name(),
      location: `${faker.location.city()}, ${faker.location.state()}`,
      rating: faker.number.int({ min: 1, max: 5 }),
      roomsAvailable: 100,
    };

    try {
      const response = await axios.post(hotelUrl, hotelData);
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

  while (flightIndex < 1) {
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
  }
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
  var totalRequestCount = process.argv[2];
  var requestCount;
  if (!totalRequestCount) {
    requestCount = 40_000;
    console.log("Generating with default RequestCount: ", requestCount);
  } else {
    if(totalRequestCount % 400 != 0) {
      console.error("Total request count must be a multiple of 400.");
      console.log("node generate-test-mass.js <TotalRequestCount>");
      process.exit(1);
    } 
    requestCount = totalRequestCount;
  }

  for (let i = 0; i < requestCount; i += 400) {
    let flightIds = await createFlights();
    let hotelIds = await createHotels();

    //  Gerar arquivos CSV dividindo o cadastro de bookings entre os usuários
    fs.rmdirSync("batches", { recursive: true });
    fs.mkdirSync("batches");

    var defaultBatchSize = 100;
    var batches = generateBatches(defaultBatchSize, flightIds, hotelIds);
    batches.forEach(saveBatch);
  }
})();
