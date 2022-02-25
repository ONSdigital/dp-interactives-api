var databases = [
    {
        name: "interactives",
        collections: ["metadata"]
    }
];

for (database of databases) {
    temp = db.getSiblingDB(database.name);
    for (collection of database.collections){
        temp.createCollection(collection);
    }
}