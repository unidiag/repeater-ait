package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/cesbo/go-mpegts"
	"github.com/cesbo/go-mpegts/crc32"
)

func aitRepeater(udpAddrStr, hbblink string) {
	ut := 0

	aitPID := uint16(0)

	// ██████╗ ██╗  ██╗     ██████╗ ██████╗ ███╗   ██╗███╗   ██╗
	// ██╔══██╗╚██╗██╔╝    ██╔════╝██╔═══██╗████╗  ██║████╗  ██║
	// ██████╔╝ ╚███╔╝     ██║     ██║   ██║██╔██╗ ██║██╔██╗ ██║
	// ██╔══██╗ ██╔██╗     ██║     ██║   ██║██║╚██╗██║██║╚██╗██║
	// ██║  ██║██╔╝ ██╗    ╚██████╗╚██████╔╝██║ ╚████║██║ ╚████║
	// ╚═╝  ╚═╝╚═╝  ╚═╝     ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═══╝

	ifi, udpAddr, port, err := parseUdpAddr(udpAddrStr)
	if err != nil {
		log.Printf("Error parsing UDP address: %v", err)
		return
	}
	eth := ""
	if ifi != nil {
		eth = ifi.Name
	}
	echo("RX: udp://" + eth + "@" + udpAddr + ":" + toStr(port))

	conn, err := openSocket4(ifi, net.ParseIP(udpAddr), port)
	if err != nil {
		log.Printf("Error opening socket: %v", err)
		return
	}
	defer conn.Close()

	// ████████╗██╗  ██╗     ██████╗ ██████╗ ███╗   ██╗███╗   ██╗██████╗
	// ╚══██╔══╝╚██╗██╔╝    ██╔════╝██╔═══██╗████╗  ██║████╗  ██║╚════██╗
	//    ██║    ╚███╔╝     ██║     ██║   ██║██╔██╗ ██║██╔██╗ ██║ █████╔╝
	//    ██║    ██╔██╗     ██║     ██║   ██║██║╚██╗██║██║╚██╗██║██╔═══╝
	//    ██║   ██╔╝ ██╗    ╚██████╗╚██████╔╝██║ ╚████║██║ ╚████║███████╗
	//    ╚═╝   ╚═╝  ╚═╝     ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝

	conn2, err := openSocket4(ifi, net.ParseIP(udpAddr), port+1)
	if err != nil {
		log.Printf("Error opening socket for transmission: %v", err)
		return
	}
	defer conn2.Close()

	destAddr := &net.UDPAddr{
		IP:   net.ParseIP(udpAddr),
		Port: port + 1,
	}

	echo("TX: udp://" + eth + "@" + udpAddr + ":" + toStr(destAddr.Port))
	echo("LINK: " + hbblink)

	// ██╗   ██╗██████╗ ██████╗      █████╗ ███╗   ██╗ █████╗ ██╗  ██╗   ██╗███████╗███████╗
	// ██║   ██║██╔══██╗██╔══██╗    ██╔══██╗████╗  ██║██╔══██╗██║  ╚██╗ ██╔╝╚══███╔╝██╔════╝
	// ██║   ██║██║  ██║██████╔╝    ███████║██╔██╗ ██║███████║██║   ╚████╔╝   ███╔╝ █████╗
	// ██║   ██║██║  ██║██╔═══╝     ██╔══██║██║╚██╗██║██╔══██║██║    ╚██╔╝   ███╔╝  ██╔══╝
	// ╚██████╔╝██████╔╝██║         ██║  ██║██║ ╚████║██║  ██║███████╗██║   ███████╗███████╗
	//  ╚═════╝ ╚═════╝ ╚═╝         ╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝╚═╝   ╚══════╝╚══════╝

	var slicer mpegts.Slicer
	buf := make([]byte, 32*1024)
	continuityCounter := byte(0)
	packetCount := 0
	interval := 1000 // Каждые 1000 пакетов добавляем AIT

	pmtPID := uint16(0)

	for {

		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			log.Printf("Error reading from socket: %v", err)
			continue
		}

		for packet := slicer.Begin(buf[:n]); packet != nil; packet = slicer.Next() {
			if packet.PID() == 0 {
				pmtPID, err = ParsePAT(packet)
				if err != nil {
					log.Println("Unknown PMT in " + udpAddr)
					return
				}
			} else if packet.PID() == mpegts.PID(pmtPID) {
				tt := int(unixtime())
				//fmt.Printf("%02X\n", packet)
				pids, _ := ParsePMT(packet) // получаем все пиды таблицы PMT
				//if aitPID == 0 || slices.Contains(pids, aitPID) { // проверяем 0 и чтобы не входило в состав имеющихся пидов
				maxPid, _ := GetMaxUint16(pids) // выделяем максимальный из них
				if len(os.Args) == 4 {
					aitPID = uint16(strToInt(os.Args[3])) // если пид был передан в третьем параметре
				} else {
					aitPID = maxPid + 101 // если пид AIT не задан на входе в функцию, то сделаем его +1 от максимального..
				}
				if ut == 0 {
					echo("AIT pid: " + toStr(aitPID))
				} else if ut != tt {
					fmt.Printf(".")
				}
				ut = tt
				//}
				packet = addNewPIDToTSPacket(packet, aitPID)
			}
			packetCount++
			conn2.WriteTo(packet, destAddr) // Отправляем основной поток как есть
			if packetCount%interval == 0 {
				aitTable := createAITTable(hbblink)
				aitPacket := createAITPacket(aitTable, aitPID, &continuityCounter)
				//fmt.Printf(">%02X\n", aitPacket)
				conn2.WriteTo(aitPacket, destAddr) // Отправляем AIT-пакет в поток
			}
		}
	}
}

//
//
//
//
//
//
//
//
// ███████╗██╗   ██╗███╗   ██╗ ██████╗
// ██╔════╝██║   ██║████╗  ██║██╔════╝
// █████╗  ██║   ██║██╔██╗ ██║██║
// ██╔══╝  ██║   ██║██║╚██╗██║██║
// ██║     ╚██████╔╝██║ ╚████║╚██████╗
// ╚═╝      ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝

func calculateCRC32(data []byte) uint32 {
	return crc32.Checksum(0xFFFFFFFF, data)
}

func createAITTable(hbbtvLink string) []byte {
	data := make([]byte, 0, 128)
	data = append(data, 0x74, 0xF0, 0x00) // Заголовок таблицы AIT
	data = append(data, 0x00, 0x10)       // application_type (HbbTV = 0x0010)
	data = append(data, 0xC3)             // version_number, current_next_indicator
	data = append(data, 0x00, 0x00)       // section_number, last_section_number
	data = append(data, 0xF0, 0x00)       // Длина списка приложений
	linkBytes := []byte(hbbtvLink)
	data = append(data, 0xF0, byte(len(linkBytes)+44))                                                    // Длина приложения (0x56 в астре, 0x57 в tvhost)
	data = append(data, 0x00, 0x00, 0x00, 0x0A)                                                           // ID приложения
	data = append(data, 0x00, 0x01, 0x01)                                                                 // Версия приложения
	data = append(data, 0xf0, 0x4d, 0x02, byte(len(linkBytes)+5), 0x00, 0x03, 0x03, byte(len(linkBytes))) // подсмотрел в астре
	data = append(data, linkBytes...)
	appname := []byte("Repeater AIT")
	// это тоже из астры...
	data = append(data, 0x00, 0x00, 0x09, 0x05, 0x00, 0x00, 0x01, 0x01, 0x01, 0xff, 0x01, 0x03, 0x01, byte(len(appname)+4), 0x65, 0x6e, 0x67, byte(len(appname)))
	data = append(data, appname...)
	sectionLength := len(data) - 3 + 4
	data[1] |= byte(sectionLength >> 8)
	data[2] = byte(sectionLength & 0xFF)
	crc := calculateCRC32(data)
	data = append(data, byte(crc>>24), byte(crc>>16), byte(crc>>8), byte(crc))
	return data
}

func createAITPacket(aitTable []byte, pid uint16, continuityCounter *byte) []byte {
	packet := make([]byte, 188)
	packet[0] = 0x47
	packet[1] = byte(0x40 | (pid >> 8))
	packet[2] = byte(pid & 0xFF)
	packet[3] = 0x10 | (*continuityCounter & 0x0F)
	*continuityCounter = (*continuityCounter + 1) % 16
	packet[4] = 0x00
	copy(packet[5:], aitTable)
	if len(aitTable) < 183 {
		for i := 5 + len(aitTable); i < 188; i++ {
			packet[i] = 0xFF
		}
	}
	//echo(packet)
	return packet
}

// ParsePAT принимает байтовый массив MPEG-TS пакета PAT и возвращает PID для PMT.
func ParsePAT(patPacket []byte) (uint16, error) {
	// Проверка длины пакета
	if len(patPacket) < 12 {
		return 0, errors.New("длина пакета слишком мала для PAT")
	}
	// Проверка синхробайта
	if patPacket[0] != 0x47 {
		return 0, errors.New("синхробайт отсутствует или некорректен")
	}
	// Извлечение флага начала полезной нагрузки (payload_unit_start_indicator)
	payloadStart := patPacket[1] & 0x40
	if payloadStart == 0 {
		return 0, errors.New("payload_unit_start_indicator не установлен")
	}
	// Указатель начала секции (pointer_field)
	pointerField := patPacket[4]
	// Смещение до начала секции
	offset := 5 + int(pointerField)
	if offset >= len(patPacket) {
		return 0, errors.New("неверное смещение до начала секции")
	}
	// Секция начинается после pointer_field
	section := patPacket[offset:]
	// Проверяем, что это таблица PAT (Table ID = 0x00)
	tableID := section[0]
	if tableID != 0x00 {
		return 0, errors.New("table ID не соответствует PAT (ожидался 0x00)")
	}
	// Извлекаем длину секции (12 бит)
	sectionLength := binary.BigEndian.Uint16(section[1:3]) & 0x0FFF
	if int(sectionLength)+3 > len(section) {
		return 0, errors.New("размер секции превышает доступную длину")
	}
	// Извлекаем Transport Stream ID (2 байта)
	transportStreamID := binary.BigEndian.Uint16(section[3:5])
	_ = transportStreamID
	// Проверка версии и текущей актуальности секции
	versionNumber := (section[5] >> 1) & 0x1F
	_ = versionNumber
	currentNextIndicator := section[5] & 0x01
	if currentNextIndicator == 0 {
		return 0, errors.New("current_next_indicator равен 0, секция неактуальна")
	}
	// Количество программ в секции
	sectionData := section[8 : 3+int(sectionLength)-4] // Секция без CRC32
	if len(sectionData) < 4 {
		return 0, errors.New("недостаточно данных для программы")
	}
	var programNumber uint16
	var pmtPID uint16
	// Парсинг всех записей о программах
	for i := 0; i < len(sectionData); i += 4 {
		if i+4 > len(sectionData) {
			return 0, errors.New("неверное количество байтов для записи о программе")
		}
		// Program Number (2 байта)
		programNumber = binary.BigEndian.Uint16(sectionData[i : i+2])
		// Program Map PID (13 бит)
		pmtPID = binary.BigEndian.Uint16(sectionData[i+2:i+4]) & 0x1FFF
		// Если программа не является сетевой PID 0, то это PMT PID
		if programNumber != 0 {
			return pmtPID, nil
		}
	}
	return 0, errors.New("не удалось найти PID для PMT")
}

// ParsePMT принимает байтовый массив MPEG-TS пакета PMT и возвращает массив PID-ов из таблицы.
func ParsePMT(pmtPacket []byte) ([]uint16, error) {
	var pids []uint16
	// Проверка длины пакета
	if len(pmtPacket) < 12 {
		return nil, errors.New("длина пакета слишком мала для PMT")
	}
	// Проверка синхробайта
	if pmtPacket[0] != 0x47 {
		return nil, errors.New("синхробайт отсутствует или некорректен")
	}
	// Проверка флага начала полезной нагрузки (payload_unit_start_indicator)
	payloadStart := pmtPacket[1] & 0x40
	if payloadStart == 0 {
		return nil, errors.New("payload_unit_start_indicator не установлен")
	}
	// Указатель начала секции (pointer_field)
	pointerField := pmtPacket[4]
	// Смещение до начала секции
	offset := 5 + int(pointerField)
	if offset >= len(pmtPacket) {
		return nil, errors.New("неверное смещение до начала секции")
	}
	// Секция начинается после pointer_field
	section := pmtPacket[offset:]
	// Проверяем, что это таблица PMT (Table ID = 0x02)
	tableID := section[0]
	if tableID != 0x02 {
		return nil, errors.New("table ID не соответствует PMT (ожидался 0x02)")
	}
	// Извлекаем длину секции (12 бит)
	sectionLength := binary.BigEndian.Uint16(section[1:3]) & 0x0FFF
	if int(sectionLength)+3 > len(section) {
		return nil, errors.New("размер секции превышает доступную длину")
	}
	// Извлекаем Program Number (2 байта)
	programNumber := binary.BigEndian.Uint16(section[3:5])
	_ = programNumber
	// Проверка версии и текущей актуальности секции
	versionNumber := (section[5] >> 1) & 0x1F
	_ = versionNumber
	currentNextIndicator := section[5] & 0x01
	if currentNextIndicator == 0 {
		return nil, errors.New("current_next_indicator равен 0, секция неактуальна")
	}
	// PCR PID (13 бит)
	pcrPID := binary.BigEndian.Uint16(section[8:10]) & 0x1FFF
	pids = append(pids, pcrPID)
	// Длина Program Info
	programInfoLength := binary.BigEndian.Uint16(section[10:12]) & 0x0FFF
	programInfoEnd := 12 + int(programInfoLength)
	if programInfoEnd > len(section) {
		return nil, errors.New("программа содержит некорректное поле длины Program Info")
	}
	// Смещение до начала информации о потоках
	streamInfoStart := programInfoEnd
	streamInfoEnd := 3 + int(sectionLength) - 4 // Конец до CRC32
	// Парсинг всех потоков
	for i := streamInfoStart; i < streamInfoEnd; {
		if i+5 > len(section) {
			return nil, errors.New("недостаточно байтов для потока в PMT")
		}
		// Stream Type (1 байт)
		streamType := section[i]
		_ = streamType
		// Elementary PID (13 бит)
		elementaryPID := binary.BigEndian.Uint16(section[i+1:i+3]) & 0x1FFF
		pids = append(pids, elementaryPID)
		// ES Info Length (12 бит)
		esInfoLength := binary.BigEndian.Uint16(section[i+3:i+5]) & 0x0FFF
		// Смещаемся к следующему описанию потока
		i += 5 + int(esInfoLength)
	}
	return pids, nil
}

// GetMaxUint16 принимает массив uint16 и возвращает максимальное значение.
func GetMaxUint16(numbers []uint16) (uint16, error) {
	if len(numbers) == 0 {
		return 0, errors.New("array empty")
	}
	maxValue := numbers[0]
	for _, num := range numbers {
		if num > maxValue {
			maxValue = num
		}
	}
	return maxValue, nil
}

// addNewPIDToTSPacket добавляет новый PID для приватных данных в полный TS-пакет (188 байт)
func addNewPIDToTSPacket(tsPacket []byte, pid uint16) []byte {
	// Смещение полезной нагрузки
	payloadStart := 4
	if tsPacket[3]&0x20 != 0 {
		adaptationFieldLength := int(tsPacket[4])
		payloadStart += 1 + adaptationFieldLength
	}
	// Учитываем pointer field
	payload := tsPacket[payloadStart:]
	pointerField := int(payload[0])
	payload = payload[1+pointerField:]
	// Получаем PMT-секцию
	sectionLength := int(binary.BigEndian.Uint16(payload[1:3]) & 0x0FFF)
	sectionData := payload[:3+sectionLength]
	streamsEnd := len(sectionData) - 4 // Без учёта CRC32

	// **Добавляем новый PID и дескриптор**
	newStream := []byte{
		0x05,                       // stream_type = 0x05 (private data)
		byte((pid>>8)&0x1F) | 0xE0, // PID (старшие 5 бит)
		byte(pid & 0xFF),           // PID (младшие 8 бит)
		0xF0, 0x05,                 // Длина дескриптора = 5 байт
		0x6F, 0x03, 0x00, 0x10, 0xE1, // Дескриптор для HbbTV
	}
	// Вставляем новый поток после всех существующих потоков
	newSectionData := append(sectionData[:streamsEnd], newStream...)
	newSectionLength := len(newSectionData) - 3 + 4 // +4 для CRC32
	binary.BigEndian.PutUint16(newSectionData[1:3], uint16(newSectionLength)|0xB000)
	//color.Yellow("%02X\n", newSectionData)
	// **Пересчитываем CRC32**
	crc := make([]byte, 4)
	binary.BigEndian.PutUint32(crc, calculateCRC32(newSectionData))
	newSectionData = append(newSectionData, crc...)
	// **Обновляем TS-пакет**
	updatedTSPacket := make([]byte, 188)
	copy(updatedTSPacket, tsPacket)
	payloadStart = payloadStart + 1 - pointerField
	copy(updatedTSPacket[payloadStart:], newSectionData)
	return updatedTSPacket
}
