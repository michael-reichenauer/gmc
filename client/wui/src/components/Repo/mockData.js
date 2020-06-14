
function createData(subject, author, datetime) {
    return {subject, author, datetime}
}

export const mockRepo = [
    createData('msg1', 'michael reichenauer', '2020-06-14 08:15'),
    createData('msg2', 'michael ', '2020-06-14 08:14'),
    createData('msg3', 'reichenauer', '2020-06-12 08:00'),
    createData('msg4', 'michel reichenauer', '2020-06-02 07:15'),
]


