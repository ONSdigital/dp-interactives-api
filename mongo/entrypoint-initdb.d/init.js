var databases = [
    {
        name: "interactives-api",
        collections: ["interactives"]
    }
];

for (database of databases) {
    temp = db.getSiblingDB(database.name);
    for (collection of database.collections){
        temp.createCollection(collection);
    }
}